package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/zahidhasann88/api-gateway/internal/config"
)

// JWTAuthMiddleware creates a middleware for JWT authentication
func JWTAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if auth is enabled
		if !cfg.Auth.Enabled {
			c.Next()
			return
		}

		// Get the auth header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		// Parse the JWT token
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		token, err := parseToken(tokenString, cfg.Auth.JWTSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Check if token is valid
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Add claims to context
			c.Set("userID", claims["sub"])
			c.Set("roles", claims["roles"])
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}
	}
}

// parseToken validates the token with the given secret
func parseToken(tokenString, secret string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
}

// GenerateToken creates a new JWT token
func GenerateToken(userID string, roles []string, cfg *config.Config) (string, error) {
	// Create claims
	expTime, _ := time.ParseDuration(cfg.Auth.Expiration)
	claims := jwt.MapClaims{
		"sub":   userID,
		"roles": roles,
		"iss":   cfg.Auth.Issuer,
		"exp":   time.Now().Add(expTime).Unix(),
		"iat":   time.Now().Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Auth.JWTSecret))
}

// AuthorizationMiddleware verifies user roles
func AuthorizationMiddleware(serviceName string, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if auth is enabled
		if !cfg.Auth.Enabled {
			c.Next()
			return
		}

		// Get service config
		serviceConfig, exists := cfg.Services[serviceName]
		if !exists || !serviceConfig.Authentication {
			c.Next()
			return
		}

		// Check if authorization is required
		if len(serviceConfig.Authorization.Roles) == 0 {
			c.Next()
			return
		}

		// Get user roles from context
		userRoles, exists := c.Get("roles")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "No role information available"})
			return
		}

		// Check if user has required role
		hasRole := false
		roles, ok := userRoles.([]interface{})
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid role format"})
			return
		}

		for _, role := range roles {
			userRole, ok := role.(string)
			if !ok {
				continue
			}
			
			for _, requiredRole := range serviceConfig.Authorization.Roles {
				if userRole == requiredRole {
					hasRole = true
					break
				}
			}
			
			if hasRole {
				break
			}
		}

		if !hasRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}

		c.Next()
	}
}