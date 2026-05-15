package lti

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"rankode/internal/config"
	"rankode/internal/models"
	db "rankode/internal/repository"
	"rankode/internal/services/users"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	LtiRoleInstructor = "http://purl.imsglobal.org/vocab/lis/v2/membership#Instructor"
)

type LtiClaims struct {
	jwt.RegisteredClaims
	DeploymentID string   `json:"https://purl.imsglobal.org/spec/lti/claim/deployment_id"`
	MessageType  string   `json:"https://purl.imsglobal.org/spec/lti/claim/message_type"`
	Version      string   `json:"https://purl.imsglobal.org/spec/lti/claim/version"`
	Roles        []string `json:"https://purl.imsglobal.org/spec/lti/claim/roles"`
	GivenName    string   `json:"given_name"`
	FamilyName   string   `json:"family_name"`
	Email        string   `json:"email"`
	Name         string   `json:"name"`
}

type LtiService struct {
	cfg          *config.Config
	q            db.Querier
	usersService *users.UserService
	jwksCache    map[string]*rsa.PublicKey
	cacheMtx     sync.RWMutex
}

func NewLtiService(cfg *config.Config, q db.Querier, usersService *users.UserService) *LtiService {
	return &LtiService{
		cfg:          cfg,
		q:            q,
		usersService: usersService,
		jwksCache:    make(map[string]*rsa.PublicKey),
	}
}

func (s *LtiService) VerifyToken(tokenString string) (*LtiClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &LtiClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("kid header not found")
		}

		return s.getPublicKey(kid)
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*LtiClaims); ok && token.Valid {
		if claims.Issuer != s.cfg.LTIIssuer {
			return nil, fmt.Errorf("invalid issuer: %s", claims.Issuer)
		}
		foundAud := false
		for _, aud := range claims.Audience {
			if aud == s.cfg.LTIClientID {
				foundAud = true
				break
			}
		}
		if !foundAud {
			return nil, fmt.Errorf("invalid audience")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (s *LtiService) getPublicKey(kid string) (*rsa.PublicKey, error) {
	s.cacheMtx.RLock()
	pubKey, ok := s.jwksCache[kid]
	s.cacheMtx.RUnlock()

	if ok {
		return pubKey, nil
	}

	// Fetch JWKS
	resp, err := http.Get(s.cfg.LTIJWKSURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var jwks struct {
		Keys []struct {
			Kid string   `json:"kid"`
			N   string   `json:"n"`
			E   string   `json:"e"`
			Kty string   `json:"kty"`
			Use string   `json:"use"`
			Alg string   `json:"alg"`
			X5c []string `json:"x5c"`
		} `json:"keys"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}

	s.cacheMtx.Lock()
	defer s.cacheMtx.Unlock()

	for _, key := range jwks.Keys {
		if key.Kty == "RSA" {
			nBytes, _ := base64.RawURLEncoding.DecodeString(key.N)
			eBytes, _ := base64.RawURLEncoding.DecodeString(key.E)
			if len(eBytes) < 4 {
				padded := make([]byte, 4)
				copy(padded[4-len(eBytes):], eBytes)
				eBytes = padded
			}
			var e int
			for _, b := range eBytes {
				e = (e << 8) | int(b)
			}

			pubKey := &rsa.PublicKey{
				N: new(big.Int).SetBytes(nBytes),
				E: e,
			}
			s.jwksCache[key.Kid] = pubKey
		}
	}

	pubKey, ok = s.jwksCache[kid]
	if !ok {
		return nil, fmt.Errorf("key not found: %s", kid)
	}

	return pubKey, nil
}

func (s *LtiService) ProvisionUser(ctx context.Context, claims *LtiClaims) (db.User, error) {
	// 1. Check if LTI link exists
	ltiUser, err := s.q.GetLtiUser(ctx, db.GetLtiUserParams{
		LtiSubject: claims.Subject,
		LtiIssuer:  claims.Issuer,
	})

	if err == nil {
		// Found existing link
		return s.q.GetUserById(ctx, ltiUser.UserID)
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return db.User{}, err
	}

	// 2. Provision new user
	username := claims.Name
	if username == "" {
		username = fmt.Sprintf("%s %s", claims.GivenName, claims.FamilyName)
	}
	if strings.TrimSpace(username) == "" {
		username = claims.Email
	}

	// Ensure unique username (simplistic)
	baseUsername := username
	for i := 1; i < 10; i++ {
		_, err := s.q.GetUsersByEmailOrUsername(ctx, username)
		if err != nil {
			break
		}
		username = fmt.Sprintf("%s_%d", baseUsername, i)
	}

	// Create user
	// We use a dummy password since it's LTI-only
	password := fmt.Sprintf("lti_%d", time.Now().UnixNano())
	isInstructor := false
	for _, role := range claims.Roles {
		if role == LtiRoleInstructor {
			isInstructor = true
			break
		}
	}
	role := int32(0)
	if isInstructor {
		role = 3
	}

	user, err := s.usersService.Register(ctx, models.CreateUserDTO{
		Username: username,
		Email:    claims.Email,
		Password: password,
		Roles:    role,
	})

	if err != nil {
		return db.User{}, err
	}

	// 4. Create LTI link
	err = s.q.CreateLtiLink(ctx, db.CreateLtiLinkParams{
		UserID:          user.ID,
		LtiSubject:      claims.Subject,
		LtiIssuer:       claims.Issuer,
		LtiDeploymentID: pgtype.Text{String: claims.DeploymentID, Valid: true},
	})

	if err != nil {
		return db.User{}, err
	}

	return user, nil
}
