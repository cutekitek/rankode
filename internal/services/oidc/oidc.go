package oidc

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	apierror "rankode/internal/errors"
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

type Claims struct {
	jwt.RegisteredClaims
	Email             string   `json:"email"`
	EmailVerified     bool     `json:"email_verified"`
	Name              string   `json:"name"`
	PreferredUsername string   `json:"preferred_username"`
	GivenName         string   `json:"given_name"`
	FamilyName        string   `json:"family_name"`
	Groups            []string `json:"groups"`
	Roles             []string `json:"roles"`
	Nonce             string   `json:"nonce"`
}

type Service struct {
	q            db.Querier
	usersService *users.UserService
	httpClient   *http.Client
	jwksCache    map[string]*rsa.PublicKey
	cacheMtx     sync.RWMutex
}

type Provider struct {
	Name                 string
	Issuer               string
	ClientID             string
	ClientSecret         string
	AuthURL              string
	TokenURL             string
	JWKSURL              string
	RedirectURL          string
	FrontendRedirectURL  string
	Scopes               []string
	AllowedDomains       []string
	RequireEmailVerified bool
}

type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	IDToken          string `json:"id_token"`
	RefreshToken     string `json:"refresh_token"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func NewOIDCService(q db.Querier, usersService *users.UserService) *Service {
	return &Service{
		q:            q,
		usersService: usersService,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		jwksCache:    make(map[string]*rsa.PublicKey),
	}
}

func (s *Service) GetProvider(ctx context.Context, name string) (Provider, error) {
	if strings.TrimSpace(name) == "" {
		name = "default"
	}
	provider, err := s.q.GetOidcProvider(ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Provider{}, apierror.WrapErrorApi(errors.New("oidc provider is not configured"), 503)
		}
		return Provider{}, err
	}

	secret := ""
	if provider.ClientSecret.Valid {
		secret = provider.ClientSecret.String
	}
	frontendRedirectURL := provider.FrontendRedirectUrl
	if frontendRedirectURL == "" {
		frontendRedirectURL = "/"
	}

	return Provider{
		Name:                 provider.Name,
		Issuer:               provider.Issuer,
		ClientID:             provider.ClientID,
		ClientSecret:         secret,
		AuthURL:              provider.AuthUrl,
		TokenURL:             provider.TokenUrl,
		JWKSURL:              provider.JwksUrl,
		RedirectURL:          provider.RedirectUrl,
		FrontendRedirectURL:  frontendRedirectURL,
		Scopes:               provider.Scopes,
		AllowedDomains:       provider.AllowedDomains,
		RequireEmailVerified: provider.RequireEmailVerified,
	}, nil
}

func (s *Service) Authenticate(ctx context.Context, provider Provider, code, codeVerifier, expectedNonce string) (db.User, error) {
	tokens, err := s.exchangeCode(ctx, provider, code, codeVerifier)
	if err != nil {
		return db.User{}, err
	}
	if tokens.IDToken == "" {
		return db.User{}, apierror.WrapErrorApi(errors.New("oidc provider did not return id_token"), 401)
	}

	claims, err := s.verifyIDToken(ctx, provider, tokens.IDToken, expectedNonce)
	if err != nil {
		return db.User{}, apierror.WrapErrorApi(err, 401)
	}

	if err := s.validateClaims(provider, claims); err != nil {
		return db.User{}, err
	}

	return s.provisionUser(ctx, provider, claims)
}

func (s *Service) exchangeCode(ctx context.Context, provider Provider, code, codeVerifier string) (tokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", provider.RedirectURL)
	form.Set("client_id", provider.ClientID)
	form.Set("code_verifier", codeVerifier)
	if provider.ClientSecret != "" {
		form.Set("client_secret", provider.ClientSecret)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, provider.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return tokenResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return tokenResponse{}, err
	}
	defer resp.Body.Close()

	var tokens tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return tokenResponse{}, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if tokens.ErrorDescription != "" {
			return tokenResponse{}, apierror.WrapErrorApi(errors.New(tokens.ErrorDescription), 401)
		}
		if tokens.Error != "" {
			return tokenResponse{}, apierror.WrapErrorApi(errors.New(tokens.Error), 401)
		}
		return tokenResponse{}, apierror.WrapErrorApi(fmt.Errorf("oidc token endpoint returned %d", resp.StatusCode), 401)
	}

	return tokens, nil
}

func (s *Service) verifyIDToken(ctx context.Context, provider Provider, tokenString, expectedNonce string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("kid header not found")
		}

		return s.getPublicKey(ctx, provider, kid)
	}, jwt.WithIssuer(provider.Issuer), jwt.WithAudience(provider.ClientID), jwt.WithExpirationRequired())
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid id_token")
	}
	if claims.Nonce == "" || claims.Nonce != expectedNonce {
		return nil, errors.New("invalid nonce")
	}
	return claims, nil
}

func (s *Service) getPublicKey(ctx context.Context, provider Provider, kid string) (*rsa.PublicKey, error) {
	cacheKey := provider.Name + ":" + kid
	s.cacheMtx.RLock()
	pubKey, ok := s.jwksCache[cacheKey]
	s.cacheMtx.RUnlock()
	if ok {
		return pubKey, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, provider.JWKSURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("oidc jwks endpoint returned %d", resp.StatusCode)
	}

	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
			Kty string `json:"kty"`
			Use string `json:"use"`
			Alg string `json:"alg"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}

	s.cacheMtx.Lock()
	defer s.cacheMtx.Unlock()

	for _, key := range jwks.Keys {
		if key.Kty != "RSA" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
		if err != nil {
			continue
		}
		if len(eBytes) < 4 {
			padded := make([]byte, 4)
			copy(padded[4-len(eBytes):], eBytes)
			eBytes = padded
		}
		var e int
		for _, b := range eBytes {
			e = (e << 8) | int(b)
		}
		s.jwksCache[provider.Name+":"+key.Kid] = &rsa.PublicKey{
			N: new(big.Int).SetBytes(nBytes),
			E: e,
		}
	}

	pubKey, ok = s.jwksCache[cacheKey]
	if !ok {
		return nil, fmt.Errorf("key not found: %s", kid)
	}
	return pubKey, nil
}

func (s *Service) validateClaims(provider Provider, claims *Claims) error {
	if claims.Subject == "" {
		return apierror.WrapErrorApi(errors.New("oidc subject is empty"), 401)
	}
	if claims.Email == "" {
		return apierror.WrapErrorApi(errors.New("oidc email claim is required"), 403)
	}
	if provider.RequireEmailVerified && !claims.EmailVerified {
		return apierror.WrapErrorApi(errors.New("oidc email is not verified"), 403)
	}
	if !emailDomainAllowed(claims.Email, provider.AllowedDomains) {
		return apierror.WrapErrorApi(errors.New("email domain is not allowed"), 403)
	}
	return nil
}

func emailDomainAllowed(email string, allowedDomains []string) bool {
	if len(allowedDomains) == 0 {
		return true
	}
	at := strings.LastIndex(email, "@")
	if at < 0 || at == len(email)-1 {
		return false
	}
	emailDomain := strings.ToLower(strings.TrimSpace(email[at+1:]))
	for _, domain := range allowedDomains {
		if emailDomain == strings.ToLower(strings.TrimSpace(domain)) {
			return true
		}
	}
	return false
}

func (s *Service) provisionUser(ctx context.Context, provider Provider, claims *Claims) (db.User, error) {
	identity, err := s.q.GetOidcIdentity(ctx, db.GetOidcIdentityParams{
		ProviderName: provider.Name,
		Subject:      claims.Subject,
	})
	if err == nil {
		return s.q.GetUserById(ctx, identity.UserID)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return db.User{}, err
	}

	user, err := s.findOrCreateUser(ctx, claims)
	if err != nil {
		return db.User{}, err
	}

	_, err = s.q.UpsertOidcIdentity(ctx, db.UpsertOidcIdentityParams{
		UserID:        user.ID,
		ProviderName:  provider.Name,
		Subject:       claims.Subject,
		Email:         pgtype.Text{String: claims.Email, Valid: claims.Email != ""},
		EmailVerified: claims.EmailVerified,
	})
	if err != nil {
		return db.User{}, err
	}

	return user, nil
}

func (s *Service) findOrCreateUser(ctx context.Context, claims *Claims) (db.User, error) {
	if claims.Email != "" {
		user, err := s.q.GetUsersByEmailOrUsername(ctx, claims.Email)
		if err == nil {
			return user, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, err
		}
	}

	username, err := s.uniqueUsername(ctx, claims)
	if err != nil {
		return db.User{}, err
	}

	return s.usersService.Register(ctx, models.CreateUserDTO{
		Username: username,
		Email:    claims.Email,
		Password: syntheticPassword(claims.Subject),
		Roles:    roleFromClaims(claims),
	})
}

func (s *Service) uniqueUsername(ctx context.Context, claims *Claims) (string, error) {
	username := claims.PreferredUsername
	if strings.TrimSpace(username) == "" {
		username = claims.Name
	}
	if strings.TrimSpace(username) == "" {
		username = strings.Split(claims.Email, "@")[0]
	}
	username = strings.TrimSpace(username)

	baseUsername := username
	for i := 0; i < 20; i++ {
		candidate := baseUsername
		if i > 0 {
			candidate = fmt.Sprintf("%s_%d", baseUsername, i)
		}
		_, err := s.q.GetUsersByEmailOrUsername(ctx, candidate)
		if errors.Is(err, pgx.ErrNoRows) {
			return candidate, nil
		}
		if err != nil {
			return "", err
		}
	}

	hash := sha256.Sum256([]byte(claims.Issuer + ":" + claims.Subject))
	return fmt.Sprintf("oidc_%s", hex.EncodeToString(hash[:])[:12]), nil
}

func syntheticPassword(subject string) string {
	hash := sha256.Sum256([]byte(subject + ":" + time.Now().String()))
	return "oidc_" + hex.EncodeToString(hash[:])
}

func roleFromClaims(claims *Claims) int32 {
	for _, value := range append(claims.Roles, claims.Groups...) {
		normalized := strings.ToLower(value)
		if strings.Contains(normalized, "teacher") || strings.Contains(normalized, "instructor") {
			return 3
		}
	}
	return 0
}
