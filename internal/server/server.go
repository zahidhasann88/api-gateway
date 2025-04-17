package server

import (
    "context"
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/zahidhasann88/api-gateway/internal/config"
    "github.com/zahidhasann88/api-gateway/pkg/logger"
)

type Server struct {
    router *gin.Engine
    server *http.Server
    config *config.Config
    logger logger.Logger  // This is unexported
}

// New creates a new server instance
func New(cfg *config.Config, log logger.Logger) *Server {
    // Set gin mode based on config
    if cfg.LogLevel != "debug" {
        gin.SetMode(gin.ReleaseMode)
    }
    
    router := gin.New()
    
    server := &http.Server{
        Addr:         cfg.Server.Address,
        Handler:      router,
        ReadTimeout:  time.Duration(cfg.Proxy.ReadTimeout) * time.Second,
        WriteTimeout: time.Duration(cfg.Proxy.WriteTimeout) * time.Second,
        IdleTimeout:  time.Duration(cfg.Proxy.IdleTimeout) * time.Second,
    }
    
    return &Server{
        router: router,
        server: server,
        config: cfg,
        logger: log,
    }
}

// Logger returns the server logger
func (s *Server) Logger() logger.Logger {
    return s.logger
}

// GET is a shortcut for router.GET
func (s *Server) GET(path string, handlers ...gin.HandlerFunc) {
    s.router.GET(path, handlers...)
}

// POST is a shortcut for router.POST
func (s *Server) POST(path string, handlers ...gin.HandlerFunc) {
    s.router.POST(path, handlers...)
}

// PUT is a shortcut for router.PUT
func (s *Server) PUT(path string, handlers ...gin.HandlerFunc) {
    s.router.PUT(path, handlers...)
}

// DELETE is a shortcut for router.DELETE
func (s *Server) DELETE(path string, handlers ...gin.HandlerFunc) {
    s.router.DELETE(path, handlers...)
}

// PATCH is a shortcut for router.PATCH
func (s *Server) PATCH(path string, handlers ...gin.HandlerFunc) {
    s.router.PATCH(path, handlers...)
}

// HEAD is a shortcut for router.HEAD
func (s *Server) HEAD(path string, handlers ...gin.HandlerFunc) {
    s.router.HEAD(path, handlers...)
}

// OPTIONS is a shortcut for router.OPTIONS
func (s *Server) OPTIONS(path string, handlers ...gin.HandlerFunc) {
    s.router.OPTIONS(path, handlers...)
}

// Any registers a route that matches all HTTP methods
func (s *Server) Any(path string, handlers ...gin.HandlerFunc) {
    s.router.Any(path, handlers...)
}

// Use attaches middleware to the router
func (s *Server) Use(middleware gin.HandlerFunc) {
    s.router.Use(middleware)
}

// Group creates a new router group
func (s *Server) Group(path string, handlers ...gin.HandlerFunc) *gin.RouterGroup {
    return s.router.Group(path, handlers...)
}

// Start starts the HTTP server
func (s *Server) Start() error {
    return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
    return s.server.Shutdown(ctx)
}