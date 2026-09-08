package transport

import (
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/ai/application"
	identity "github.com/ksamwang/PersonalContentPlatform/internal/identity/transport"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/httpx"
	"net/http"
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
		r.Use(h.auth)
		r.Post("/v1/workspaces/{workspaceID}/ai/suggestions", h.suggest)
		r.Post("/v1/workspaces/{workspaceID}/ai/suggestions/{suggestionID}:review", h.review)
	})
}
func principal(r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	p, ok := identity.Principal(r.Context())
	ws, err := uuid.Parse(chi.URLParam(r, "workspaceID"))
	return p.WorkspaceID, p.UserID, ok && err == nil && ws == p.WorkspaceID
}

type suggestInput struct {
	TargetID uuid.UUID `json:"target_id"`
	Purpose  string    `json:"purpose"`
	Input    string    `json:"input"`
}

func (h *HTTP) suggest(w http.ResponseWriter, r *http.Request) {
	ws, user, ok := principal(r)
	if !ok {
		httpx.Error(w, 403, "workspace_forbidden", "workspace access denied")
		return
	}
	var in suggestInput
	if err := httpx.Decode(w, r, &in); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return
	}
	idValue, err := h.service.Suggest(r.Context(), ws, user, in.TargetID, in.Purpose, in.Input)
	if err != nil {
		httpx.Error(w, 503, "ai_unavailable", err.Error())
		return
	}
	httpx.JSON(w, 201, map[string]any{"suggestion_id": idValue, "state": "pending"})
}

type reviewInput struct {
	State string `json:"state"`
}

func (h *HTTP) review(w http.ResponseWriter, r *http.Request) {
	ws, user, ok := principal(r)
	if !ok {
		httpx.Error(w, 403, "workspace_forbidden", "workspace access denied")
		return
	}
	idValue, err := uuid.Parse(chi.URLParam(r, "suggestionID"))
	var in reviewInput
	if err != nil || httpx.Decode(w, r, &in) != nil {
		httpx.Error(w, 400, "invalid_request", "invalid review request")
		return
	}
	if err = h.service.Review(r.Context(), ws, user, idValue, in.State); err != nil {
		httpx.Error(w, 422, "review_invalid", err.Error())
		return
	}
	w.WriteHeader(204)
}
