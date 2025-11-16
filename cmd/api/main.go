package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/T1mof/pr-reviewer-service/internal/config"
	"github.com/T1mof/pr-reviewer-service/internal/handler"
	"github.com/T1mof/pr-reviewer-service/internal/repository"
	"github.com/T1mof/pr-reviewer-service/internal/service"
)

func main() {
	config.SetupLogger()

	if err := run(); err != nil {
		slog.Error("Fatal error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	slog.Info("Starting PR Reviewer Service...")

	cfg := config.Load()

	db, err := cfg.ConnectDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("Failed to close database connection", "error", err)
		}
	}()

	if err := cfg.RunMigrations(db.DB); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	repo := repository.NewRepository(db.DB)
	svc := service.NewReviewerService(repo)
	h := handler.NewHandler(svc, cfg.AdminToken)

	srv := startServer(cfg.Port, h.SetupRouter())

	waitForShutdown(srv)

	return nil
}

// startServer запускает HTTP сервер.
func startServer(port string, handler http.Handler) *http.Server {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("Server is starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	return srv
}

// waitForShutdown ожидает сигнал остановки и gracefully завершает сервер.
func waitForShutdown(srv *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exited gracefully")
}
