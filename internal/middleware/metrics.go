// Create a new file internal/middleware/metrics.go
package middleware

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_gateway_requests_total",
			Help: "Total number of requests",
		},
		[]string{"service", "method", "status"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_gateway_request_duration_seconds",
			Help:    "Duration of requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method"},
	)
)

// Metrics middleware collects metrics about requests
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Get service name from path
		service := "unknown"
		path := c.Request.URL.Path

		// Extract service name from path
		// Assuming format: /api/{service}/...
		if len(path) > 5 && path[:5] == "/api/" {
			parts := strings.Split(path[5:], "/")
			if len(parts) > 0 {
				service = parts[0]
			}
		}

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()

		requestCount.WithLabelValues(service, c.Request.Method, status).Inc()
		requestDuration.WithLabelValues(service, c.Request.Method).Observe(duration)
	}
}
