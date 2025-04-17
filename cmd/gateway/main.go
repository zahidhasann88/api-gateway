package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zahidhasann88/api-gateway/internal/config"
	"github.com/zahidhasann88/api-gateway/internal/handlers"
	"github.com/zahidhasann88/api-gateway/internal/middleware"
	"github.com/zahidhasann88/api-gateway/internal/server"
	"github.com/zahidhasann88/api-gateway/pkg/logger"
)

func main() {
	// Initialize configuration
	cfg, err := config.Load("./configs")
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	// Initialize logger
	log := logger.New(cfg.LogLevel)
	defer log.Sync()

	// Initialize the server
	srv := server.New(cfg, log)

	// Register middlewares
	srv.Use(middleware.RequestID())
	srv.Use(middleware.Logger(log))
	srv.Use(middleware.Recovery(log))
	srv.Use(middleware.CORS(cfg.CORS))

	// Register routes
	handlers.RegisterRoutes(srv, cfg)

	// Start the server in a goroutine
	go func() {
		log.Info("Starting API Gateway server", "address", cfg.Server.Address)
		if err := srv.Start(); err != nil {
			log.Error("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
	}

	log.Info("Server exited properly")
}
