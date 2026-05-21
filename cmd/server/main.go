package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chessgoddess/chesslens/internal/api"
	"github.com/chessgoddess/chesslens/internal/auth"
	"github.com/chessgoddess/chesslens/internal/config"
	"github.com/chessgoddess/chesslens/internal/db"
)

func main() {
	// Structured JSON logging in production, text in development.
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if cfg.Environment == "production" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	}

	ctx := context.Background()

	database, err := db.New(ctx, cfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	authService := auth.NewService(cfg)

	srv := api.New(cfg, database, authService)

	var httpServer *http.Server
	if cfg.HTTPEnabled {
		httpServer = &http.Server{
			Addr:         ":" + cfg.Port,
			Handler:      srv.Handler(),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		go func() {
			slog.Info("chesslens server starting", "port", cfg.Port, "env", cfg.Environment)
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("server error", "error", err)
				os.Exit(1)
			}
		}()
	}

	if !cfg.HTTPEnabled && !cfg.WorkerEnabled {
		slog.Error("both HTTP and worker disabled, nothing to do")
		os.Exit(1)
	}

	if !cfg.HTTPEnabled {
		slog.Info("running in worker-only mode")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if httpServer != nil {
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("forced shutdown", "error", err)
		}
	}

	srv.Stop()
	slog.Info("server exited")
}
