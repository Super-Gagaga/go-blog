package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
	"github.com/yourname/go-blog/internal/pkg/slogx"
)

const TraceIDKey = "trace_id"

func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = newTraceID()
		}
		c.Set(TraceIDKey, traceID)
		c.Header("X-Trace-ID", traceID)
		// 同步进 request context,使缓存/数据访问层的日志也携带 trace_id
		c.Request = c.Request.WithContext(slogx.WithRequestID(c.Request.Context(), traceID))
		c.Next()
	}
}

func GetTraceID(c *gin.Context) string {
	value, ok := c.Get(TraceIDKey)
	if !ok {
		return ""
	}
	traceID, _ := value.(string)
	return traceID
}

func newTraceID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "trace-fallback"
	}
	return hex.EncodeToString(buf)
}
