package main

import (
	"context"
	"fmt"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/config"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/database"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/migrate"
	"log/slog"
	"os"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] != "up" {
		fmt.Fprintln(os.Stderr, "usage: migrate up")
		os.Exit(2)
	}
	ctx := context.Background()
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
	if err = migrate.Up(ctx, db); err != nil {
		slog.Error("migrate", "error", err)
		os.Exit(1)
	}
	slog.Info("database is current")
}
