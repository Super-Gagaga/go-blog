package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"github.com/yourname/go-blog/internal/config"
)

func newTestManager() *Manager {
	return NewManager(config.JWTConfig{
		Issuer:        "go-blog-test",
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    7 * 24 * time.Hour,
	})
}

func TestManagerAccessTokenRoundTrip(t *testing.T) {
	manager := newTestManager()
	token, err := manager.GenerateAccessToken(42)
	require.NoError(t, err)

	claims, err := manager.ParseAccessToken(token)
	require.NoError(t, err)
	require.Equal(t, int64(42), claims.UserID)
	require.Equal(t, TokenTypeAccess, claims.TokenType)
	require.Equal(t, "go-blog-test", claims.Issuer)
	require.False(t, claims.ExpiresAt.Time.Before(time.Now()))
}

func TestManagerRejectsCrossTypeTokens(t *testing.T) {
	manager := newTestManager()

	refresh, err := manager.GenerateRefreshToken(42)
	require.NoError(t, err)
	_, err = manager.ParseAccessToken(refresh)
	require.Error(t, err)

	access, err := manager.GenerateAccessToken(42)
	require.NoError(t, err)
	_, err = manager.ParseRefreshToken(access)
	require.Error(t, err)
}

func TestManagerRejectsWrongSecret(t *testing.T) {
	manager := newTestManager()
	token, err := manager.GenerateAccessToken(42)
	require.NoError(t, err)

	other := NewManager(config.JWTConfig{
		Issuer:       "go-blog-test",
		AccessSecret: "another-secret",
		AccessTTL:    time.Minute,
	})
	_, err = other.ParseAccessToken(token)
	require.Error(t, err)
}

func TestManagerRejectsExpiredToken(t *testing.T) {
	manager := NewManager(config.JWTConfig{
		Issuer:       "go-blog-test",
		AccessSecret: "test-access-secret",
		AccessTTL:    -time.Minute,
	})
	token, err := manager.GenerateAccessToken(42)
	require.NoError(t, err)

	_, err = manager.ParseAccessToken(token)
	require.Error(t, err)
}

func TestManagerRejectsTokenWithoutExpiry(t *testing.T) {
	unsigned := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    int64(7),
		"token_type": TokenTypeAccess,
		"iss":        "go-blog-test",
	})
	token, err := unsigned.SignedString([]byte("test-access-secret"))
	require.NoError(t, err)

	_, err = newTestManager().ParseAccessToken(token)
	require.Error(t, err)
}
