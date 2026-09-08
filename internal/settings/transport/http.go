package transport

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	identity "github.com/ksamwang/PersonalContentPlatform/internal/identity/transport"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/httpx"
	"github.com/ksamwang/PersonalContentPlatform/internal/settings/application"
	"github.com/ksamwang/PersonalContentPlatform/internal/settings/domain"
)

type HTTP struct {
	service *application.Service
	auth    func(http.Handler) http.Handler
}

func NewHTTP(s *application.Service, a func(http.Handler) http.Handler) *HTTP {
	return &HTTP{service: s, auth: a}
}
func (h *HTTP) Register(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(h.auth, h.scope)
		r.Get("/v1/workspaces/{workspaceID}/settings", h.get)
		r.Put("/v1/workspaces/{workspaceID}/settings/{section}", h.updateSection)
		r.Put("/v1/workspaces/{workspaceID}/settings/storage", h.saveStorage)
		r.Put("/v1/workspaces/{workspaceID}/settings/ai", h.saveAI)
		r.Post("/v1/workspaces/{workspaceID}/settings/storage:test", h.testStorage)
		r.Post("/v1/workspaces/{workspaceID}/settings/ai:test", h.testAI)
	})
}

func (h *HTTP) testStorage(w http.ResponseWriter, r *http.Request) {
	if !owner(w, r) {
		return
	}
	ws, _, _ := principal(r)
	if err := h.service.TestStorage(r.Context(), ws); err != nil {
		httpx.Error(w, 422, "storage_connection_failed", err.Error())
		return
	}
	httpx.JSON(w, 200, map[string]bool{"ok": true})
}
func (h *HTTP) testAI(w http.ResponseWriter, r *http.Request) {
	if !owner(w, r) {
		return
	}
	ws, _, _ := principal(r)
	if err := h.service.TestAI(r.Context(), ws); err != nil {
		httpx.Error(w, 422, "ai_connection_failed", err.Error())
		return
	}
	httpx.JSON(w, 200, map[string]bool{"ok": true})
}
func (h *HTTP) scope(next http.Handler) http.Handler {
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
func principal(r *http.Request) (uuid.UUID, uuid.UUID, string) {
	p, _ := identity.Principal(r.Context())
	return p.WorkspaceID, p.UserID, p.Role
}
func owner(w http.ResponseWriter, r *http.Request) bool {
	_, _, role := principal(r)
	if role != "owner" {
		httpx.Error(w, 403, "owner_required", "only workspace owners can change settings")
		return false
	}
	return true
}
func (h *HTTP) get(w http.ResponseWriter, r *http.Request) {
	ws, _, role := principal(r)
	v, err := h.service.Get(r.Context(), ws, role == "owner")
	if err != nil {
		httpx.Error(w, 500, "settings_failed", "could not load settings")
		return
	}
	httpx.JSON(w, 200, v)
}
func (h *HTTP) updateSection(w http.ResponseWriter, r *http.Request) {
	if !owner(w, r) {
		return
	}
	section := chi.URLParam(r, "section")
	if section == "storage" || section == "ai" {
		httpx.Error(w, 400, "invalid_section", "use the provider settings endpoint")
		return
	}
	var raw json.RawMessage
	if err := httpx.Decode(w, r, &raw); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return
	}
	ws, user, _ := principal(r)
	v, err := h.service.UpdateGeneral(r.Context(), ws, user, section, raw)
	if err != nil {
		httpx.Error(w, 422, "settings_invalid", err.Error())
		return
	}
	httpx.JSON(w, 200, v)
}
func (h *HTTP) saveStorage(w http.ResponseWriter, r *http.Request) {
	if !owner(w, r) {
		return
	}
	var v domain.StorageProfile
	if err := httpx.Decode(w, r, &v); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return
	}
	ws, user, _ := principal(r)
	saved, err := h.service.SaveStorage(r.Context(), ws, user, v)
	if err != nil {
		httpx.Error(w, 422, "storage_invalid", err.Error())
		return
	}
	httpx.JSON(w, 200, saved)
}
func (h *HTTP) saveAI(w http.ResponseWriter, r *http.Request) {
	if !owner(w, r) {
		return
	}
	var v domain.AIConfig
	if err := httpx.Decode(w, r, &v); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return
	}
	ws, user, _ := principal(r)
	saved, err := h.service.SaveAI(r.Context(), ws, user, v)
	if err != nil {
		httpx.Error(w, 422, "ai_invalid", err.Error())
		return
	}
	httpx.JSON(w, 200, saved)
}
