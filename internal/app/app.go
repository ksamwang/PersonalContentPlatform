package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	identityapp "github.com/ksamwang/PersonalContentPlatform/internal/identity/application"
	identitypg "github.com/ksamwang/PersonalContentPlatform/internal/identity/infrastructure/postgres"
	identityhttp "github.com/ksamwang/PersonalContentPlatform/internal/identity/transport"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/config"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/health"
	"net/http"
	"time"
)

type App struct {
	DB        *pgxpool.Pool
	StartedAt time.Time
	identity  *identityhttp.HTTP
}

func New(db *pgxpool.Pool, cfg config.Config) (*App, error) {
	repo := identitypg.New(db)
	sessions := identityapp.NewService(repo, cfg.PasswordPepper, cfg.SessionTTL)
	passkeys, err := identityapp.NewPasskeys(repo, cfg.WebAuthnRPID, cfg.WebAuthnOrigins)
	if err != nil {
		return nil, err
	}
	return &App{DB: db, StartedAt: time.Now().UTC(), identity: identityhttp.NewHTTP(sessions, passkeys, cfg.SessionCookieName, cfg.SessionTTL, cfg.Environment != "development")}, nil
}
func (a *App) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, middleware.Timeout(30*time.Second))
	health.RegisterRoutes(r, a.DB)
	a.identity.Register(r)
	return r
}
