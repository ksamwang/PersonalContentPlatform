package main

import (
	"context"
	integration "github.com/ksamwang/PersonalContentPlatform/internal/integration/application"
	integrationpg "github.com/ksamwang/PersonalContentPlatform/internal/integration/infrastructure/postgres"
	webhook "github.com/ksamwang/PersonalContentPlatform/internal/integration/infrastructure/webhook"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/config"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/database"
	publication "github.com/ksamwang/PersonalContentPlatform/internal/publication/application"
	retrieval "github.com/ksamwang/PersonalContentPlatform/internal/retrieval/application"
	settingspg "github.com/ksamwang/PersonalContentPlatform/internal/settings/infrastructure/postgres"
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
	integrationRepository := integrationpg.New(db)
	errors := make(chan error, 2)
	search := retrieval.New(db, settingspg.New(db))
	go func() { errors <- publication.NewProcessor(db, "default").WithIndexer(search).Run(ctx) }()
	go func() {
		errors <- integration.NewWebhookDispatcher(integrationRepository, webhook.NewSender(), 8).Run(ctx)
	}()
	if err := <-errors; err != nil {
		slog.Error("worker failed", "error", err)
		os.Exit(1)
	}
	slog.Info("worker stopped")
}
