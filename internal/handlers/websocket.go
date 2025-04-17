package handlers

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/zahidhasann88/api-gateway/internal/config"
	"github.com/zahidhasann88/api-gateway/pkg/logger"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	config   *config.Config
	logger   logger.Logger
	upgrader websocket.Upgrader
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(cfg *config.Config, log logger.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		config: cfg,
		logger: log,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins (in production you should restrict this)
			},
		},
	}
}

// ProxyWebSocket handles WebSocket connections
func (h *WebSocketHandler) ProxyWebSocket(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if service exists
		serviceConfig, exists := h.config.Services[serviceName]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
			return
		}

		// Parse the target URL
		targetURL, err := url.Parse(serviceConfig.URL)
		if err != nil {
			h.logger.Error("Failed to parse WebSocket URL", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Modify the target URL for WebSocket
		wsScheme := "ws"
		if strings.HasPrefix(targetURL.Scheme, "https") {
			wsScheme = "wss"
		}
		wsURL := url.URL{
			Scheme:   wsScheme,
			Host:     targetURL.Host,
			Path:     c.Request.URL.Path,
			RawQuery: c.Request.URL.RawQuery,
		}

		// Log the connection attempt
		h.logger.Info("WebSocket connection attempt",
			"service", serviceName,
			"client", c.ClientIP(),
			"target", wsURL.String())

		// Upgrade the HTTP connection to WebSocket
		conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			h.logger.Error("Failed to upgrade connection", "error", err)
			return
		}
		defer conn.Close()

		// Connect to the backend WebSocket
		backendConn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
		if err != nil {
			h.logger.Error("Failed to connect to backend WebSocket", "error", err)
			conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Cannot connect to service"))
			return
		}
		defer backendConn.Close()

		// Create channels for done signal
		clientDone := make(chan struct{})
		backendDone := make(chan struct{})

		// Forward messages from client to backend
		go func() {
			defer close(clientDone)
			for {
				messageType, message, err := conn.ReadMessage()
				if err != nil {
					return
				}
				err = backendConn.WriteMessage(messageType, message)
				if err != nil {
					return
				}
			}
		}()

		// Forward messages from backend to client
		go func() {
			defer close(backendDone)
			for {
				messageType, message, err := backendConn.ReadMessage()
				if err != nil {
					return
				}
				err = conn.WriteMessage(messageType, message)
				if err != nil {
					return
				}
			}
		}()

		// Wait for either connection to close
		select {
		case <-clientDone:
			backendConn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		case <-backendDone:
			conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		}
	}
}
