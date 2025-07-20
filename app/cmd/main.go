package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kentech-project/internal/adapters/http/server"
	"kentech-project/pkg/config"
	"kentech-project/pkg/database"
	"kentech-project/pkg/logger"
	"kentech-project/pkg/tracing"
)

func main() {
	logger := logger.New()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := database.NewPostgresConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Errorf("Failed to close database connection", "error", err)
		} else {
			logger.Info("Database connection closed")
		}
	}(db)

	shutdown, err := tracing.InitTracer("kentech-project")
	if err != nil {
		log.Fatalf("Failed to initialize tracing: %v", err)
	}
	defer shutdown(context.Background())

	serverInstance := server.NewServer(cfg, db, logger)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: serverInstance.Handler(),
	}

	go func() {
		logger.Info("Starting server on port " + cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}
