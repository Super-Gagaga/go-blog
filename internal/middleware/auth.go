package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yourname/go-blog/internal/pkg/app"
	"github.com/yourname/go-blog/internal/pkg/errcode"
	jwtpkg "github.com/yourname/go-blog/internal/pkg/jwt"
)

const ContextUserIDKey = "user_id"

func Auth(jwt *jwtpkg.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "missing token")
			c.Abort()
			return
		}

		claims, err := jwt.ParseAccessToken(token)
		if err != nil {
			app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "invalid token")
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Next()
	}
}

func CurrentUserID(c *gin.Context) (int64, bool) {
	value, ok := c.Get(ContextUserIDKey)
	if !ok {
		return 0, false
	}

	userID, ok := value.(int64)
	return userID, ok
}

func extractToken(c *gin.Context) string {
	if token := strings.TrimSpace(c.Query("token")); token != "" {
		return token
	}

	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}

	return authHeader
}
