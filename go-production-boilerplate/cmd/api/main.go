package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/saurav11sarkar/go-production-boilerplate/internal/app"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/config"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/database"
)

func main() {
	cfg := config.MustLoad()
	level := slog.LevelInfo
	if cfg.Environment == "development" {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level, AddSource: cfg.Environment == "development"}))
	slog.SetDefault(logger)

	db, err := database.Connect(context.Background(), cfg)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	application, err := app.New(cfg, logger, db)
	if err != nil {
		logger.Error("application setup failed", "error", err)
		os.Exit(1)
	}
	if err := application.Run(); err != nil {
		logger.Error("application stopped", "error", err)
		os.Exit(1)
	}
}
