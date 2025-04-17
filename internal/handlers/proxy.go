package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/zahidhasann88/api-gateway/internal/config"
	"github.com/zahidhasann88/api-gateway/pkg/logger"
)

type ProxyHandler struct {
	config    *config.Config
	logger    logger.Logger
	limiters  map[string]*rate.Limiter
	transport http.RoundTripper
}

func NewProxyHandler(cfg *config.Config, log logger.Logger) *ProxyHandler {
	limiters := make(map[string]*rate.Limiter)

	// Initialize rate limiters for each service
	for serviceName, serviceConfig := range cfg.Services {
		limiter := rate.NewLimiter(rate.Limit(serviceConfig.RateLimit), serviceConfig.RateLimit)
		limiters[serviceName] = limiter
	}

	transport := &http.Transport{
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	return &ProxyHandler{
		config:    cfg,
		logger:    log,
		limiters:  limiters,
		transport: transport,
	}
}

func (h *ProxyHandler) ProxyRequest(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceConfig, exists := h.config.Services[serviceName]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
			return
		}

		// Check rate limiting
		limiter := h.limiters[serviceName]
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}

		// Parse the target URL
		targetURL, err := url.Parse(serviceConfig.URL)
		if err != nil {
			h.logger.Error("Failed to parse service URL",
				"service", serviceName,
				"url", serviceConfig.URL,
				"error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid service configuration"})
			return
		}

		// Create the reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		// Set custom transport with timeout
		proxy.Transport = &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			DisableKeepAlives:     false,
			ResponseHeaderTimeout: time.Duration(serviceConfig.Timeout) * time.Second,
		}

		// Set director to modify the request
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)

			// Copy the request ID
			if requestID, exists := c.Get("RequestID"); exists {
				req.Header.Set("X-Request-ID", fmt.Sprintf("%v", requestID))
			}

			// Forward the original path
			req.URL.Path = c.Request.URL.Path
			if c.Request.URL.RawQuery != "" {
				req.URL.RawQuery = c.Request.URL.RawQuery
			}

			// Add gateway headers
			req.Header.Set("X-Gateway-Service", serviceName)
			req.Header.Set("X-Forwarded-For", c.ClientIP())

			h.logger.Debug("Proxying request",
				"service", serviceName,
				"method", req.Method,
				"path", req.URL.Path)
		}

		// Modify the response
		proxy.ModifyResponse = func(resp *http.Response) error {
			// Add response headers
			resp.Header.Set("X-Gateway-Service", serviceName)

			// Log response status
			h.logger.Debug("Received response",
				"service", serviceName,
				"status", resp.StatusCode)

			return nil
		}

		// Handle errors
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			h.logger.Error("Proxy error",
				"service", serviceName,
				"error", err)

			c.JSON(http.StatusBadGateway, gin.H{
				"error": "Service unavailable",
			})
		}

		// Save the original response writer
		originalWriter := c.Writer

		// Create a response recorder to capture the response
		responseRecorder := &responseRecorder{
			ResponseWriter: originalWriter,
			Body:           new(bytes.Buffer),
		}

		c.Writer = responseRecorder

		// Serve the request through proxy
		proxy.ServeHTTP(c.Writer, c.Request)

		// Restore the original writer
		c.Writer = originalWriter
	}
}

// responseRecorder is a custom ResponseWriter that captures the response body
type responseRecorder struct {
	gin.ResponseWriter
	Body *bytes.Buffer
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.Body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r *responseRecorder) WriteString(s string) (int, error) {
	r.Body.WriteString(s)
	return r.ResponseWriter.WriteString(s)
}
