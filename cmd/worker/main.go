package main

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	aiapp "github.com/ksamwang/PersonalContentPlatform/internal/ai/application"
	openai "github.com/ksamwang/PersonalContentPlatform/internal/ai/infrastructure/openai"
	assetapp "github.com/ksamwang/PersonalContentPlatform/internal/asset/application"
	assetpg "github.com/ksamwang/PersonalContentPlatform/internal/asset/infrastructure/postgres"
	"github.com/ksamwang/PersonalContentPlatform/internal/asset/infrastructure/storagefactory"
	inboxapp "github.com/ksamwang/PersonalContentPlatform/internal/inbox/application"
	inboxpg "github.com/ksamwang/PersonalContentPlatform/internal/inbox/infrastructure/postgres"
	integration "github.com/ksamwang/PersonalContentPlatform/internal/integration/application"
	integrationpg "github.com/ksamwang/PersonalContentPlatform/internal/integration/infrastructure/postgres"
	webhook "github.com/ksamwang/PersonalContentPlatform/internal/integration/infrastructure/webhook"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/config"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/database"
	publication "github.com/ksamwang/PersonalContentPlatform/internal/publication/application"
	retrieval "github.com/ksamwang/PersonalContentPlatform/internal/retrieval/application"
	settingspg "github.com/ksamwang/PersonalContentPlatform/internal/settings/infrastructure/postgres"
	taskapp "github.com/ksamwang/PersonalContentPlatform/internal/task/application"
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
	settingsRepository := settingspg.New(db)
	search := retrieval.New(db, settingsRepository)
	fallbackStorage, err := storagefactory.New(ctx, cfg)
	if err != nil {
		slog.Error("configure storage", "error", err)
		os.Exit(1)
	}
	assets := assetapp.New(assetpg.New(db), storagefactory.NewResolver(settingsRepository, fallbackStorage, cfg.StorageBasePath))
	inbox := inboxapp.New(inboxpg.New(db), cfg.SupportedLocales, settingsRepository, assets)
	ai := aiapp.New(db, openai.NewResolver(settingsRepository, openai.New(cfg.AIBaseURL, cfg.AIAPIKey, cfg.AIModel)))
	handlers := map[string]taskapp.Handler{
		"inbox.process": func(ctx context.Context, ws uuid.UUID, raw json.RawMessage) (string, error) {
			var input struct {
				ItemID uuid.UUID `json:"item_id"`
			}
			if err := json.Unmarshal(raw, &input); err != nil {
				return "", err
			}
			item, err := inbox.Process(ctx, ws, input.ItemID)
			if err != nil {
				return "", err
			}
			return taskapp.ResultID(item.ID), nil
		},
		"ai.suggest": func(ctx context.Context, ws uuid.UUID, raw json.RawMessage) (string, error) {
			var input struct {
				UserID         uuid.UUID `json:"user_id"`
				TargetID       uuid.UUID `json:"target_id"`
				LocalizationID uuid.UUID `json:"localization_id"`
				Purpose        string    `json:"purpose"`
				Input          string    `json:"input"`
			}
			if err := json.Unmarshal(raw, &input); err != nil {
				return "", err
			}
			suggestionID, err := ai.Suggest(ctx, ws, input.UserID, input.TargetID, input.LocalizationID, input.Purpose, input.Input)
			if err != nil {
				return "", err
			}
			return taskapp.ResultID(suggestionID), nil
		},
		"ai.translate": func(ctx context.Context, ws uuid.UUID, raw json.RawMessage) (string, error) {
			var input struct {
				UserID               uuid.UUID `json:"user_id"`
				SourceLocalizationID uuid.UUID `json:"source_localization_id"`
				TargetLocalizationID uuid.UUID `json:"target_localization_id"`
			}
			if err := json.Unmarshal(raw, &input); err != nil {
				return "", err
			}
			draft, err := ai.Translate(ctx, ws, input.UserID, input.SourceLocalizationID, input.TargetLocalizationID)
			if err != nil {
				return "", err
			}
			return taskapp.ResultID(draft.LocalizationID), nil
		},
	}
	errors := make(chan error, 3)
	cacheInvalidator := publication.NewCacheInvalidator(cfg.PublicWebOrigin, cfg.PublicRevalidateToken)
	go func() {
		errors <- publication.NewProcessor(db, "default").WithIndexer(search).WithCache(cacheInvalidator).Run(ctx)
	}()
	go func() { errors <- taskapp.NewProcessor(db, "default", handlers).Run(ctx) }()
	go func() {
		errors <- integration.NewWebhookDispatcher(integrationRepository, webhook.NewSender(), 8).Run(ctx)
	}()
	if err := <-errors; err != nil {
		slog.Error("worker failed", "error", err)
		os.Exit(1)
	}
	slog.Info("worker stopped")
}
