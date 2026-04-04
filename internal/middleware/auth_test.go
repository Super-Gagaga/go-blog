package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/yourname/go-blog/internal/config"
	jwtpkg "github.com/yourname/go-blog/internal/pkg/jwt"
)

func newTestJWTManager(accessTTL time.Duration) *jwtpkg.Manager {
	return jwtpkg.NewManager(config.JWTConfig{
		Issuer:        "go-blog-test",
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessTTL:     accessTTL,
		RefreshTTL:    7 * 24 * time.Hour,
	})
}

func performAuth(manager *jwtpkg.Manager, token string) (*httptest.ResponseRecorder, int64, bool) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	var gotUserID int64
	var reached bool
	r.Use(Auth(manager))
	r.GET("/private", func(c *gin.Context) {
		reached = true
		gotUserID, _ = CurrentUserID(c)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w, gotUserID, reached
}

func TestAuthMiddlewareAllowsValidToken(t *testing.T) {
	manager := newTestJWTManager(15 * time.Minute)
	token, err := manager.GenerateAccessToken(42)
	require.NoError(t, err)

	w, userID, reached := performAuth(manager, token)
	require.Equal(t, http.StatusOK, w.Code)
	require.True(t, reached)
	require.Equal(t, int64(42), userID)
}

func TestAuthMiddlewareRejectsMissingToken(t *testing.T) {
	w, _, reached := performAuth(newTestJWTManager(time.Minute), "")
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.False(t, reached)
}

func TestAuthMiddlewareRejectsTamperedToken(t *testing.T) {
	manager := newTestJWTManager(time.Minute)
	token, err := manager.GenerateAccessToken(42)
	require.NoError(t, err)

	w, _, reached := performAuth(manager, token+"x")
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.False(t, reached)
}

func TestAuthMiddlewareRejectsExpiredToken(t *testing.T) {
	manager := newTestJWTManager(-time.Minute)
	token, err := manager.GenerateAccessToken(42)
	require.NoError(t, err)

	w, _, reached := performAuth(manager, token)
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.False(t, reached)
}

func TestAuthMiddlewareRejectsRefreshTokenAsAccess(t *testing.T) {
	manager := newTestJWTManager(time.Minute)
	token, err := manager.GenerateRefreshToken(42)
	require.NoError(t, err)

	w, _, reached := performAuth(manager, token)
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.False(t, reached)
}
