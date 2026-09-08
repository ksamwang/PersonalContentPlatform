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
	contentapp "github.com/ksamwang/PersonalContentPlatform/internal/content/application"
	contentpg "github.com/ksamwang/PersonalContentPlatform/internal/content/infrastructure/postgres"
	contenthttp "github.com/ksamwang/PersonalContentPlatform/internal/content/transport"
	identityapp "github.com/ksamwang/PersonalContentPlatform/internal/identity/application"
	identitypg "github.com/ksamwang/PersonalContentPlatform/internal/identity/infrastructure/postgres"
	identityhttp "github.com/ksamwang/PersonalContentPlatform/internal/identity/transport"
	integrationapp "github.com/ksamwang/PersonalContentPlatform/internal/integration/application"
	integrationpg "github.com/ksamwang/PersonalContentPlatform/internal/integration/infrastructure/postgres"
	integrationhttp "github.com/ksamwang/PersonalContentPlatform/internal/integration/transport"
	knowledgeapp "github.com/ksamwang/PersonalContentPlatform/internal/knowledge/application"
	knowledgepg "github.com/ksamwang/PersonalContentPlatform/internal/knowledge/infrastructure/postgres"
	knowledgehttp "github.com/ksamwang/PersonalContentPlatform/internal/knowledge/transport"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/config"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/health"
	publicationhttp "github.com/ksamwang/PersonalContentPlatform/internal/publication/transport"
	"net/http"
	"time"
)

type App struct {
	DB          *pgxpool.Pool
	StartedAt   time.Time
	identity    *identityhttp.HTTP
	content     *contenthttp.HTTP
	public      *publicationhttp.HTTP
	asset       *assethttp.HTTP
	knowledge   *knowledgehttp.HTTP
	ai          *aihttp.HTTP
	integration *integrationhttp.HTTP
}

func New(db *pgxpool.Pool, cfg config.Config) (*App, error) {
	repo := identitypg.New(db)
	sessions := identityapp.NewService(repo, cfg.PasswordPepper, cfg.SessionTTL)
	passkeys, err := identityapp.NewPasskeys(repo, cfg.WebAuthnRPID, cfg.WebAuthnOrigins)
	if err != nil {
		return nil, err
	}
	identityTransport := identityhttp.NewHTTP(sessions, passkeys, cfg.SessionCookieName, cfg.SessionTTL, cfg.Environment != "development")
	contentService := contentapp.New(contentpg.New(db), cfg.SupportedLocales)
	storage, err := storagefactory.New(context.Background(), cfg)
	if err != nil {
		return nil, err
	}
	assetService := assetapp.New(assetpg.New(db), storage)
	knowledgeService := knowledgeapp.New(knowledgepg.New(db))
	aiService := aiapp.New(db, openai.New(cfg.AIBaseURL, cfg.AIAPIKey, cfg.AIModel))
	integrationRepository := integrationpg.New(db)
	return &App{
		DB:          db,
		StartedAt:   time.Now().UTC(),
		identity:    identityTransport,
		content:     contenthttp.NewHTTP(contentService, identityTransport.RequireSession),
		public:      publicationhttp.NewHTTP(db),
		asset:       assethttp.NewHTTP(assetService, identityTransport.RequireSession),
		knowledge:   knowledgehttp.NewHTTP(knowledgeService, identityTransport.RequireSession),
		ai:          aihttp.NewHTTP(aiService, identityTransport.RequireSession),
		integration: integrationhttp.NewHTTP(integrationapp.NewExportService(integrationRepository), integrationapp.NewWebhookService(integrationRepository), identityTransport.RequireSession),
	}, nil
}
func (a *App) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, middleware.Timeout(30*time.Second))
	health.RegisterRoutes(r, a.DB)
	a.identity.Register(r)
	a.content.Register(r)
	a.public.Register(r)
	a.asset.Register(r)
	a.knowledge.Register(r)
	a.ai.Register(r)
	a.integration.Register(r)
	return r
}
