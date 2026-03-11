package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/yourname/go-blog/internal/pkg/app"
	"github.com/yourname/go-blog/internal/pkg/errcode"
	"go.uber.org/zap"
)

func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered",
					zap.Any("error", r),
					zap.ByteString("stack", debug.Stack()),
				)
				app.Error(c, http.StatusInternalServerError, errcode.InternalServer, "")
				c.Abort()
			}
		}()

		c.Next()
	}
}
