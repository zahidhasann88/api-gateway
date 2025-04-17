package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zahidhasann88/api-gateway/internal/config"
	"github.com/zahidhasann88/api-gateway/internal/middleware"
	"github.com/zahidhasann88/api-gateway/internal/server"
)

// handleLogin handles authentication requests
func handleLogin(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Authentication logic - this is a placeholder
		// In a real implementation, you would validate credentials and generate a token
		var loginRequest struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&loginRequest); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		// Simple validation (would be more comprehensive in production)
		if loginRequest.Username == "" || loginRequest.Password == "" {
			c.JSON(400, gin.H{"error": "Username and password required"})
			return
		}

		// Generate JWT token (simplified)
		token, err := middleware.GenerateToken(loginRequest.Username, []string{"user"}, cfg)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(200, gin.H{
			"token": token,
			"user":  loginRequest.Username,
		})
	}
}

func RegisterRoutes(srv *server.Server, cfg *config.Config) {
	// Create handlers
	proxyHandler := NewProxyHandler(cfg, srv.Logger())
	graphqlHandler := NewGraphQLHandler(cfg, srv.Logger())
	wsHandler := NewWebSocketHandler(cfg, srv.Logger())

	// Register global middleware
	srv.Use(middleware.RequestID())
	srv.Use(middleware.Logger(srv.Logger()))
	srv.Use(middleware.Recovery(srv.Logger()))
	srv.Use(middleware.CORS(cfg.CORS))
	srv.Use(middleware.Metrics())
	srv.Use(middleware.TransformationMiddleware(cfg, srv.Logger()))

	// Healthcheck endpoint
	srv.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Metrics endpoint
	srv.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Authentication endpoints
	auth := srv.Group("/auth")
	{
		auth.POST("/login", handleLogin(cfg))
	}

	// API routes
	api := srv.Group("/api")
	api.Use(middleware.JWTAuthMiddleware(cfg))

	// Create separate GraphQL and WebSocket routes to avoid conflicts with wildcards

	// GraphQL endpoints
	graphql := api.Group("/graphql")
	{
		graphql.POST("/users", graphqlHandler.HandleRequest("users"))
		graphql.POST("/payments", graphqlHandler.HandleRequest("payments"))
		graphql.POST("/orders", graphqlHandler.HandleRequest("orders"))
		// General purpose GraphQL endpoint for service aggregation
		graphql.POST("", func(c *gin.Context) {
			// Implementation would depend on your GraphQL schema aggregation strategy
			c.JSON(501, gin.H{"error": "Not implemented"})
		})
	}

	// WebSocket endpoints
	ws := api.Group("/ws")
	{
		ws.GET("/users/*path", wsHandler.ProxyWebSocket("users"))
		ws.GET("/payments/*path", wsHandler.ProxyWebSocket("payments"))
		ws.GET("/notifications/*path", wsHandler.ProxyWebSocket("notifications"))
	}

	// Service routes for standard REST APIs

	// Users service routes
	users := api.Group("/users")
	users.Use(middleware.AuthorizationMiddleware("users", cfg))
	{
		users.Any("/*path", proxyHandler.ProxyRequest("users"))
	}

	// Payments service routes
	payments := api.Group("/payments")
	payments.Use(middleware.AuthorizationMiddleware("payments", cfg))
	{
		payments.Any("/*path", proxyHandler.ProxyRequest("payments"))
	}

	// Notifications service routes
	notifications := api.Group("/notifications")
	notifications.Use(middleware.AuthorizationMiddleware("notifications", cfg))
	{
		notifications.Any("/*path", proxyHandler.ProxyRequest("notifications"))
	}

	// Orders service routes
	orders := api.Group("/orders")
	orders.Use(middleware.AuthorizationMiddleware("orders", cfg))
	{
		orders.Any("/*path", proxyHandler.ProxyRequest("orders"))
	}

	// You can add additional service routes here as needed
}
