package transport

import (
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	identity "github.com/ksamwang/PersonalContentPlatform/internal/identity/transport"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/httpx"
	"github.com/ksamwang/PersonalContentPlatform/internal/publication/application"
	"net/http"
	"time"
)

type AdminHTTP struct {
	service *application.Service
	auth    func(http.Handler) http.Handler
}

func NewAdminHTTP(s *application.Service, auth func(http.Handler) http.Handler) *AdminHTTP {
	return &AdminHTTP{s, auth}
}
func (h *AdminHTTP) Register(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(h.auth, h.scope)
		r.Get("/v1/workspaces/{workspaceID}/publication-records", h.list)
		r.Get("/v1/workspaces/{workspaceID}/publication-channels", h.channels)
		r.Post("/v1/workspaces/{workspaceID}/publication-records:schedule", h.schedule)
		r.Post("/v1/workspaces/{workspaceID}/publication-records/{publicationID}:retry", h.retry)
		r.Post("/v1/workspaces/{workspaceID}/publication-records/{publicationID}:withdraw", h.withdraw)
	})
}
func (h *AdminHTTP) scope(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := identity.Principal(r.Context())
		ws, err := uuid.Parse(chi.URLParam(r, "workspaceID"))
		if !ok || err != nil || p.WorkspaceID != ws {
			httpx.Error(w, 403, "workspace_forbidden", "workspace access denied")
			return
		}
		next.ServeHTTP(w, r)
	})
}
func adminPrincipal(r *http.Request) (uuid.UUID, uuid.UUID, string) {
	p, _ := identity.Principal(r.Context())
	return p.WorkspaceID, p.UserID, p.Role
}
func canPublish(w http.ResponseWriter, r *http.Request) bool {
	_, _, role := adminPrincipal(r)
	if role != "owner" && role != "editor" {
		httpx.Error(w, 403, "editor_required", "workspace editor access is required")
		return false
	}
	return true
}
func (h *AdminHTTP) list(w http.ResponseWriter, r *http.Request) {
	ws, _, _ := adminPrincipal(r)
	items, err := h.service.List(r.Context(), ws)
	if err != nil {
		httpx.Error(w, 500, "publication_list_failed", err.Error())
		return
	}
	httpx.JSON(w, 200, map[string]any{"items": items})
}
func (h *AdminHTTP) channels(w http.ResponseWriter, r *http.Request) {
	ws, _, _ := adminPrincipal(r)
	items, err := h.service.Channels(r.Context(), ws)
	if err != nil {
		httpx.Error(w, 500, "publication_channels_failed", err.Error())
		return
	}
	httpx.JSON(w, 200, map[string]any{"items": items})
}
func pubID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	idv, err := uuid.Parse(chi.URLParam(r, "publicationID"))
	if err != nil {
		httpx.Error(w, 400, "invalid_id", "invalid publication ID")
		return uuid.Nil, false
	}
	return idv, true
}
func (h *AdminHTTP) schedule(w http.ResponseWriter, r *http.Request) {
	if !canPublish(w, r) {
		return
	}
	var input struct {
		ContentID   uuid.UUID `json:"content_id"`
		Locale      string    `json:"locale"`
		ScheduledAt time.Time `json:"scheduled_at"`
	}
	if err := httpx.Decode(w, r, &input); err != nil || input.ScheduledAt.IsZero() {
		httpx.Error(w, 400, "invalid_request", "content_id, locale and scheduled_at are required")
		return
	}
	ws, user, _ := adminPrincipal(r)
	idv, err := h.service.Schedule(r.Context(), ws, user, input.ContentID, input.Locale, input.ScheduledAt)
	if err != nil {
		httpx.Error(w, 422, "publication_schedule_failed", err.Error())
		return
	}
	httpx.JSON(w, 202, map[string]any{"publication_id": idv})
}
func (h *AdminHTTP) retry(w http.ResponseWriter, r *http.Request) {
	if !canPublish(w, r) {
		return
	}
	idv, ok := pubID(w, r)
	if !ok {
		return
	}
	ws, _, _ := adminPrincipal(r)
	if err := h.service.Retry(r.Context(), ws, idv); err != nil {
		httpx.Error(w, 409, "publication_retry_failed", err.Error())
		return
	}
	w.WriteHeader(204)
}
func (h *AdminHTTP) withdraw(w http.ResponseWriter, r *http.Request) {
	if !canPublish(w, r) {
		return
	}
	idv, ok := pubID(w, r)
	if !ok {
		return
	}
	ws, _, _ := adminPrincipal(r)
	if err := h.service.Withdraw(r.Context(), ws, idv); err != nil {
		httpx.Error(w, 409, "publication_withdraw_failed", err.Error())
		return
	}
	w.WriteHeader(204)
}
