package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourname/go-blog/internal/pkg/app"
	"github.com/yourname/go-blog/internal/pkg/errcode"
	"github.com/yourname/go-blog/internal/repository"
)

func Admin(users *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := CurrentUserID(c)
		if !ok {
			app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
			c.Abort()
			return
		}

		user, err := users.FindByID(userID)
		if err != nil || user == nil {
			app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
			c.Abort()
			return
		}
		if user.Role != "admin" {
			app.Error(c, http.StatusForbidden, errcode.Forbidden, "admin permission required")
			c.Abort()
			return
		}

		c.Next()
	}
}
