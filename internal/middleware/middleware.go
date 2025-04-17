package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zahidhasann88/api-gateway/internal/config"
	"github.com/zahidhasann88/api-gateway/pkg/logger"
)

// RequestID adds a unique ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// Logger logs the incoming request and response details
func Logger(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		requestID, _ := c.Get("RequestID")

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		log.Info("Request processed",
			"requestID", requestID,
			"status", statusCode,
			"method", method,
			"path", path,
			"ip", clientIP,
			"latency", latency,
		)
	}
}

// Recovery handles panics and returns 500 error
func Recovery(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get("RequestID")
				log.Error("Request panic recovered",
					"requestID", requestID,
					"error", err,
				)
				c.AbortWithStatusJSON(500, gin.H{
					"error": "Internal Server Error",
				})
			}
		}()
		c.Next()
	}
}

// CORS handles Cross-Origin Resource Sharing
func CORS(config config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", config.AllowedOrigins[0])
		c.Writer.Header().Set("Access-Control-Allow-Methods", joinStrings(config.AllowedMethods))
		c.Writer.Header().Set("Access-Control-Allow-Headers", joinStrings(config.AllowedHeaders))

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func joinStrings(strings []string) string {
	result := ""
	for i, s := range strings {
		result += s
		if i < len(strings)-1 {
			result += ", "
		}
	}
	return result
}
