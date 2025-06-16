package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"rankode/internal/config"
	db "rankode/internal/repository"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	jwtSecret string
}

func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{jwtSecret: cfg.JWTSecret}
}

func (s *AuthService) GenerateToken(user db.User) (string, error) {

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "rankode",
		Subject:   strconv.Itoa(int(user.ID)),
		Audience:  jwt.ClaimStrings{"rankode-app"},
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		NotBefore: jwt.NewNumericDate(time.Now()),
	})

	tokenString, err := claims.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return tokenString, nil
}

func (s *AuthService) VerifyToken(tokenString string) (int32, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return 0, errors.New("invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims format")
	}

	subject, err := claims.GetSubject()
	if err != nil {
		return 0, errors.New("subject claim not found in token")
	}

	userID, err := strconv.Atoi(subject)
	if err != nil {
		return 0, errors.New("invalid user ID format in token")
	}

	return int32(userID), nil
}
