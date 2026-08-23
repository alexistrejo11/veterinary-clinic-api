package middleware

import (
	"time"

	"clinic-vet-api/internal/shared/log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func AuditLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		requestID := uuid.New().String()
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()

		duration := time.Since(start)
		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("client_ip", c.ClientIP()),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status_code", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.Int("error_count", len(c.Errors)),
		}
		if userID, ok := c.Get("userID"); ok {
			if id, ok := userID.(uint); ok {
				fields = append(fields, zap.Uint("user_id", id))
			}
		}
		log.Audit.Info("audit_log", fields...)

		log.App.Info("Request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
		)
	}
}
