package app

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	aiapp "github.com/ksamwang/PersonalContentPlatform/internal/ai/application"
	openai "github.com/ksamwang/PersonalContentPlatform/internal/ai/infrastructure/openai"
	aihttp "github.com/ksamwang/PersonalContentPlatform/internal/ai/transport"
	assetapp "github.com/ksamwang/PersonalContentPlatform/internal/asset/application"
	assetpg "github.com/ksamwang/PersonalContentPlatform/internal/asset/infrastructure/postgres"
	"github.com/ksamwang/PersonalContentPlatform/internal/asset/infrastructure/storagefactory"
	assethttp "github.com/ksamwang/PersonalContentPlatform/internal/asset/transport"
	collectionapp "github.com/ksamwang/PersonalContentPlatform/internal/collection/application"
	collectionpg "github.com/ksamwang/PersonalContentPlatform/internal/collection/infrastructure/postgres"
	collectionhttp "github.com/ksamwang/PersonalContentPlatform/internal/collection/transport"
	contentapp "github.com/ksamwang/PersonalContentPlatform/internal/content/application"
	contentpg "github.com/ksamwang/PersonalContentPlatform/internal/content/infrastructure/postgres"
	contenthttp "github.com/ksamwang/PersonalContentPlatform/internal/content/transport"
	identityapp "github.com/ksamwang/PersonalContentPlatform/internal/identity/application"
	identitypg "github.com/ksamwang/PersonalContentPlatform/internal/identity/infrastructure/postgres"
	identityhttp "github.com/ksamwang/PersonalContentPlatform/internal/identity/transport"
	inboxapp "github.com/ksamwang/PersonalContentPlatform/internal/inbox/application"
	inboxpg "github.com/ksamwang/PersonalContentPlatform/internal/inbox/infrastructure/postgres"
	inboxhttp "github.com/ksamwang/PersonalContentPlatform/internal/inbox/transport"
	integrationapp "github.com/ksamwang/PersonalContentPlatform/internal/integration/application"
	integrationpg "github.com/ksamwang/PersonalContentPlatform/internal/integration/infrastructure/postgres"
	integrationhttp "github.com/ksamwang/PersonalContentPlatform/internal/integration/transport"
	knowledgeapp "github.com/ksamwang/PersonalContentPlatform/internal/knowledge/application"
	knowledgepg "github.com/ksamwang/PersonalContentPlatform/internal/knowledge/infrastructure/postgres"
	knowledgehttp "github.com/ksamwang/PersonalContentPlatform/internal/knowledge/transport"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/config"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/health"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/httpx"
	publicationapp "github.com/ksamwang/PersonalContentPlatform/internal/publication/application"
	publicationhttp "github.com/ksamwang/PersonalContentPlatform/internal/publication/transport"
	retrievalapp "github.com/ksamwang/PersonalContentPlatform/internal/retrieval/application"
	retrievalhttp "github.com/ksamwang/PersonalContentPlatform/internal/retrieval/transport"
	settingsapp "github.com/ksamwang/PersonalContentPlatform/internal/settings/application"
	settingspg "github.com/ksamwang/PersonalContentPlatform/internal/settings/infrastructure/postgres"
	settingshttp "github.com/ksamwang/PersonalContentPlatform/internal/settings/transport"
	taskapp "github.com/ksamwang/PersonalContentPlatform/internal/task/application"
	taskhttp "github.com/ksamwang/PersonalContentPlatform/internal/task/transport"
	"net/http"
	"time"
)

type App struct {
	DB               *pgxpool.Pool
	StartedAt        time.Time
	identity         *identityhttp.HTTP
	content          *contenthttp.HTTP
	public           *publicationhttp.HTTP
	publicationAdmin *publicationhttp.AdminHTTP
	asset            *assethttp.HTTP
	knowledge        *knowledgehttp.HTTP
	ai               *aihttp.HTTP
	integration      *integrationhttp.HTTP
	settings         *settingshttp.HTTP
	inbox            *inboxhttp.HTTP
	collection       *collectionhttp.HTTP
	retrieval        *retrievalhttp.HTTP
	task             *taskhttp.HTTP
}

func New(db *pgxpool.Pool, cfg config.Config) (*App, error) {
	repo := identitypg.New(db)
	sessions := identityapp.NewService(repo, cfg.PasswordPepper, cfg.SessionTTL)
	passkeys, err := identityapp.NewPasskeys(repo, cfg.WebAuthnRPID, cfg.WebAuthnOrigins)
	if err != nil {
		return nil, err
	}
	identityTransport := identityhttp.NewHTTP(sessions, passkeys, cfg.SessionCookieName, cfg.SessionTTL, cfg.Environment != "development")
	fallbackStorage, err := storagefactory.New(context.Background(), cfg)
	if err != nil {
		return nil, err
	}
	settingsRepository := settingspg.New(db)
	contentService := contentapp.New(contentpg.New(db), cfg.SupportedLocales, settingsRepository)
	storage := storagefactory.NewResolver(settingsRepository, fallbackStorage, cfg.StorageBasePath)
	assetService := assetapp.New(assetpg.New(db), storage)
	knowledgeService := knowledgeapp.New(knowledgepg.New(db))
	aiService := aiapp.New(db, openai.NewResolver(settingsRepository, openai.New(cfg.AIBaseURL, cfg.AIAPIKey, cfg.AIModel)))
	taskQueue := taskapp.NewQueue(db)
	cacheInvalidator := publicationapp.NewCacheInvalidator(cfg.PublicWebOrigin, cfg.PublicRevalidateToken)
	integrationRepository := integrationpg.New(db)
	return &App{
		DB:               db,
		StartedAt:        time.Now().UTC(),
		identity:         identityTransport,
		content:          contenthttp.NewHTTP(contentService, identityTransport.RequireSession),
		public:           publicationhttp.NewHTTP(db),
		publicationAdmin: publicationhttp.NewAdminHTTP(publicationapp.NewService(db).WithCache(cacheInvalidator), identityTransport.RequireSession),
		asset:            assethttp.NewHTTP(assetService, identityTransport.RequireSession),
		knowledge:        knowledgehttp.NewHTTP(knowledgeService, identityTransport.RequireSession),
		ai:               aihttp.NewHTTP(aiService, identityTransport.RequireSession).WithQueue(taskQueue),
		integration: func() *integrationhttp.HTTP {
			exports := integrationapp.NewExportService(integrationRepository)
			return integrationhttp.NewHTTP(exports, integrationapp.NewWebhookService(integrationRepository), identityTransport.RequireSession).WithMigration(integrationapp.NewArchiveService(exports, assetService), integrationapp.NewImportService(db, assetService))
		}(),
		settings:   settingshttp.NewHTTP(settingsapp.New(settingsRepository, cfg.StorageBasePath), identityTransport.RequireSession),
		inbox:      inboxhttp.NewHTTP(inboxapp.New(inboxpg.New(db), cfg.SupportedLocales, settingsRepository, assetService), identityTransport.RequireSession).WithQueue(taskQueue),
		collection: collectionhttp.NewHTTP(collectionapp.New(collectionpg.New(db)), identityTransport.RequireSession),
		retrieval:  retrievalhttp.NewHTTP(retrievalapp.New(db, settingsRepository), identityTransport.RequireSession),
		task:       taskhttp.NewHTTP(taskQueue, identityTransport.RequireSession),
	}, nil
}
func (a *App) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, httpx.AccessLog, middleware.RealIP, middleware.Recoverer, middleware.Timeout(30*time.Second))
	health.RegisterRoutes(r, a.DB)
	a.identity.Register(r)
	a.content.Register(r)
	a.public.Register(r)
	a.publicationAdmin.Register(r)
	a.asset.Register(r)
	a.knowledge.Register(r)
	a.ai.Register(r)
	a.integration.Register(r)
	a.settings.Register(r)
	a.inbox.Register(r)
	a.collection.Register(r)
	a.retrieval.Register(r)
	if a.task != nil {
		a.task.Register(r)
	}
	return r
}
