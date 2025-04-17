package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zahidhasann88/api-gateway/internal/config"
	"github.com/zahidhasann88/api-gateway/pkg/logger"
)

// GraphQLHandler handles GraphQL requests
type GraphQLHandler struct {
	config *config.Config
	logger logger.Logger
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
}

// NewGraphQLHandler creates a new GraphQL handler
func NewGraphQLHandler(cfg *config.Config, log logger.Logger) *GraphQLHandler {
	return &GraphQLHandler{
		config: cfg,
		logger: log,
	}
}

// HandleRequest handles a GraphQL request
func (h *GraphQLHandler) HandleRequest(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if service exists
		serviceConfig, exists := h.config.Services[serviceName]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
			return
		}

		// Parse the GraphQL request
		var graphqlRequest GraphQLRequest
		if err := c.ShouldBindJSON(&graphqlRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid GraphQL request"})
			return
		}

		// Make request to the service
		requestData, err := json.Marshal(graphqlRequest)
		if err != nil {
			h.logger.Error("Failed to marshal GraphQL request", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Create request to service
		req, err := http.NewRequest("POST", serviceConfig.URL+"/graphql", bytes.NewBuffer(requestData))
		if err != nil {
			h.logger.Error("Failed to create GraphQL request", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		if requestID, exists := c.Get("RequestID"); exists {
			req.Header.Set("X-Request-ID", requestID.(string))
		}

		// Forward auth headers
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			req.Header.Set("Authorization", authHeader)
		}

		// Make the request
		client := &http.Client{
			Timeout: time.Duration(serviceConfig.Timeout) * time.Second,
		}
		resp, err := client.Do(req)
		if err != nil {
			h.logger.Error("GraphQL request failed", "error", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": "Service unavailable"})
			return
		}
		defer resp.Body.Close()

		// Read response
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			h.logger.Error("Failed to read GraphQL response", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Parse and forward response
		c.Header("Content-Type", "application/json")
		c.Status(resp.StatusCode)
		c.Writer.Write(body)
	}
}
