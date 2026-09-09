package transport

import (
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/content/application"
	identityhttp "github.com/ksamwang/PersonalContentPlatform/internal/identity/transport"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/httpx"
	"net/http"
)

type HTTP struct {
	service *application.Service
	auth    func(http.Handler) http.Handler
}

func NewHTTP(service *application.Service, auth func(http.Handler) http.Handler) *HTTP {
	return &HTTP{service: service, auth: auth}
}
func (h *HTTP) Register(r chi.Router) {
	r.Get("/v1/previews/{token}", h.preview)
	r.Group(func(r chi.Router) {
		r.Use(h.auth)
		r.Route("/v1/workspaces/{workspaceID}", func(r chi.Router) {
			r.Use(workspaceScope)
			r.Get("/contents", h.list)
			r.Post("/contents", h.create)
			r.Post("/contents/{contentID}/localizations", h.createLocalization)
			r.Get("/contents/{contentID}", h.get)
			r.Patch("/contents/{contentID}", h.updateProperties)
			r.Post("/contents/{contentID}:archive", h.archive)
			r.Post("/contents/{contentID}:restore", h.restore)
			r.Delete("/contents/{contentID}", h.remove)
			r.Put("/localizations/{localizationID}/draft", h.saveDraft)
			r.Get("/localizations/{localizationID}/draft", h.getDraft)
			r.Post("/localizations/{localizationID}/revisions", h.sealRevision)
			r.Get("/localizations/{localizationID}/revisions", h.listRevisions)
			r.Get("/localizations/{localizationID}/revisions/{revisionID}", h.getRevision)
			r.Post("/localizations/{localizationID}/revisions/{revisionID}:restore", h.restoreRevision)
			r.Post("/localizations/{localizationID}/preview", h.createPreview)
			r.Post("/localizations/{localizationID}:mark-ready", h.markReady)
			r.Get("/localizations/{localizationID}/readiness", h.readiness)
			r.Post("/publications", h.publish)
			r.Get("/search", h.search)
		})
	})
}
func workspaceScope(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := identityhttp.Principal(r.Context())
		workspaceID, err := uuid.Parse(chi.URLParam(r, "workspaceID"))
		if !ok || err != nil || workspaceID != principal.WorkspaceID {
			httpx.Error(w, http.StatusForbidden, "workspace_forbidden", "workspace access denied")
			return
		}
		next.ServeHTTP(w, r)
	})
}
func principal(r *http.Request) (uuid.UUID, uuid.UUID) {
	p, _ := identityhttp.Principal(r.Context())
	return p.WorkspaceID, p.UserID
}
func parseID(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	value, err := uuid.Parse(chi.URLParam(r, name))
	if err != nil {
		httpx.Error(w, 400, "invalid_id", "invalid "+name)
		return uuid.Nil, false
	}
	return value, true
}
