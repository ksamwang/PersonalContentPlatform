package transport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	identity "github.com/ksamwang/PersonalContentPlatform/internal/identity/transport"
	"github.com/ksamwang/PersonalContentPlatform/internal/integration/application"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/httpx"
)

type HTTP struct {
	exports  *application.ExportService
	webhooks *application.WebhookService
	auth     func(http.Handler) http.Handler
}

func NewHTTP(exports *application.ExportService, webhooks *application.WebhookService, auth func(http.Handler) http.Handler) *HTTP {
	return &HTTP{exports: exports, webhooks: webhooks, auth: auth}
}

func (h *HTTP) Register(router chi.Router) {
	router.Group(func(router chi.Router) {
		router.Use(h.auth)
		router.With(scopeWorkspace).Get("/v1/workspaces/{workspaceID}/exports/manifest.json", h.exportManifest)
		router.With(scopeWorkspace).Get("/v1/workspaces/{workspaceID}/webhooks", h.listWebhooks)
		router.With(scopeWorkspace).Post("/v1/workspaces/{workspaceID}/webhooks", h.createWebhook)
		router.With(scopeWorkspace).Patch("/v1/workspaces/{workspaceID}/webhooks/{endpointID}", h.setWebhookEnabled)
		router.With(scopeWorkspace).Delete("/v1/workspaces/{workspaceID}/webhooks/{endpointID}", h.deleteWebhook)
	})
}

func scopeWorkspace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := identity.Principal(r.Context())
		workspaceID, err := uuid.Parse(chi.URLParam(r, "workspaceID"))
		if !ok || err != nil || principal.WorkspaceID != workspaceID {
			httpx.Error(w, http.StatusForbidden, "workspace_forbidden", "workspace access denied")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func principal(r *http.Request) identityPrincipal {
	value, _ := identity.Principal(r.Context())
	return identityPrincipal{UserID: value.UserID, WorkspaceID: value.WorkspaceID}
}

type identityPrincipal struct {
	UserID      uuid.UUID
	WorkspaceID uuid.UUID
}
