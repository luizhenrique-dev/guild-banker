package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/luizhenrique-dev/guild-banker/api/config"
	"github.com/luizhenrique-dev/guild-banker/api/database"
	"github.com/luizhenrique-dev/guild-banker/api/internal/infra/webserver"
)

func main() {
	logFile, err := os.OpenFile("guildbanker.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		slog.Error("failed to open log file", "error", err)
		os.Exit(1)
	}
	defer func() {
		if closeErr := logFile.Close(); closeErr != nil {
			slog.Error("failed to close log file", "error", closeErr)
		}
	}()

	logOutput := io.MultiWriter(logFile, os.Stdout)
	logger := slog.New(slog.NewTextHandler(logOutput, nil))
	slog.SetDefault(logger)
	gin.DefaultWriter = logOutput
	gin.DefaultErrorWriter = logOutput

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()

	db, err := database.New(dbCtx, cfg.DB)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			slog.Error("failed to close database", "error", closeErr)
		}
	}()

	slog.Info("database connected",
		"host", cfg.DB.Host,
		"port", cfg.DB.Port,
		"name", cfg.DB.Name,
	)

	slog.Info("running database migrations")
	if err := database.Migrate(db, cfg.DB.Name); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("migrations applied successfully")

	server := webserver.NewServer(cfg, db)

	go func() {
		if startErr := server.Start(); startErr != nil && !errors.Is(startErr, http.ErrServerClosed) {
			slog.Error("server error", "error", startErr)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if shutdownErr := server.Shutdown(ctx); shutdownErr != nil {
		slog.Error("forced shutdown", "error", shutdownErr)
		os.Exit(1)
	}

	slog.Info("server exited gracefully")
}
