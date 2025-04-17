package handlers

import (
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    
    "github.com/zahidhasann88/api-gateway/internal/config"
    "github.com/zahidhasann88/api-gateway/pkg/logger"
)

func TestProxyHandler_ProxyRequest(t *testing.T) {
    // Setup
    gin.SetMode(gin.TestMode)
    
    // Mock server for the target service
    targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"message": "Hello from target service"}`))
    }))
    defer targetServer.Close()
    
    // Create test configuration
    cfg := &config.Config{
        Services: map[string]config.ServiceConfig{
            "test-service": {
                URL:        targetServer.URL,
                Timeout:    5,
                RetryCount: 3,
                RateLimit:  10,
            },
        },
    }
    
    log := logger.New("debug")
    
    // Create proxy handler
    proxyHandler := NewProxyHandler(cfg, log)
    
    // Create test router
    router := gin.New()
    router.Use(func(c *gin.Context) {
        c.Set("RequestID", "test-request-id")
        c.Next()
    })
    router.GET("/test/*path", proxyHandler.ProxyRequest("test-service"))
    
    // Create test request
    req := httptest.NewRequest("GET", "/test/hello", nil)
    w := httptest.NewRecorder()
    
    // Execute request
    router.ServeHTTP(w, req)
    
    // Verify response
    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Body.String(), "Hello from target service")
}