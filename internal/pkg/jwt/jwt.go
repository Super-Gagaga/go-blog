package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yourname/go-blog/internal/config"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type Manager struct {
	cfg config.JWTConfig
}

type Claims struct {
	UserID    int64  `json:"user_id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func NewManager(cfg config.JWTConfig) *Manager {
	return &Manager{cfg: cfg}
}

func (m *Manager) GenerateAccessToken(userID int64) (string, error) {
	return m.generateToken(userID, TokenTypeAccess, m.cfg.AccessTTL, []byte(m.cfg.AccessSecret))
}

func (m *Manager) GenerateRefreshToken(userID int64) (string, error) {
	return m.generateToken(userID, TokenTypeRefresh, m.cfg.RefreshTTL, []byte(m.cfg.RefreshSecret))
}

func (m *Manager) ParseAccessToken(tokenStr string) (*Claims, error) {
	return m.parseToken(tokenStr, TokenTypeAccess, []byte(m.cfg.AccessSecret))
}

func (m *Manager) ParseRefreshToken(tokenStr string) (*Claims, error) {
	return m.parseToken(tokenStr, TokenTypeRefresh, []byte(m.cfg.RefreshSecret))
}

func (m *Manager) generateToken(userID int64, tokenType string, ttl time.Duration, secret []byte) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func (m *Manager) parseToken(tokenStr string, expectedType string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	}, jwt.WithIssuer(m.cfg.Issuer))
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("invalid token type")
	}
	return claims, nil
}
