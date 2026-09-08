package main

import (
	"context"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/config"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/database"
	publication "github.com/ksamwang/PersonalContentPlatform/internal/publication/application"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}
	db, err := database.Open(ctx, cfg.DatabaseURL, cfg.DBPoolMax)
	if err != nil {
		slog.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("worker started")
	if err := publication.NewProcessor(db, "default").Run(ctx); err != nil {
		slog.Error("worker failed", "error", err)
		os.Exit(1)
	}
	slog.Info("worker stopped")
}
