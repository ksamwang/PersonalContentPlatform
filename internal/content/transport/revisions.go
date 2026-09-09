package transport

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ksamwang/PersonalContentPlatform/internal/content/application"
	identity "github.com/ksamwang/PersonalContentPlatform/internal/identity/transport"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/httpx"
)

func editorRequired(w http.ResponseWriter, r *http.Request) bool {
	p, _ := identity.Principal(r.Context())
	if p.Role != "owner" && p.Role != "editor" {
		httpx.Error(w, 403, "editor_required", "workspace editor access is required")
		return false
	}
	return true
}
func (h *HTTP) listRevisions(w http.ResponseWriter, r *http.Request) {
	localizationID, ok := parseID(w, r, "localizationID")
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	workspaceID, _ := principal(r)
	items, err := h.service.ListRevisions(r.Context(), workspaceID, localizationID, limit)
	if err != nil {
		httpx.Error(w, 500, "revisions_failed", "could not load revision history")
		return
	}
	httpx.JSON(w, 200, map[string]any{"items": items})
}
func (h *HTTP) getRevision(w http.ResponseWriter, r *http.Request) {
	localizationID, ok := parseID(w, r, "localizationID")
	if !ok {
		return
	}
	revisionID, ok := parseID(w, r, "revisionID")
	if !ok {
		return
	}
	workspaceID, _ := principal(r)
	revision, err := h.service.GetRevision(r.Context(), workspaceID, localizationID, revisionID)
	if err != nil {
		httpx.Error(w, 404, "revision_not_found", "revision was not found")
		return
	}
	httpx.JSON(w, 200, revision)
}
func (h *HTTP) restoreRevision(w http.ResponseWriter, r *http.Request) {
	if !editorRequired(w, r) {
		return
	}
	localizationID, ok := parseID(w, r, "localizationID")
	if !ok {
		return
	}
	revisionID, ok := parseID(w, r, "revisionID")
	if !ok {
		return
	}
	var input struct {
		Version int `json:"version"`
	}
	if err := httpx.Decode(w, r, &input); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return
	}
	workspaceID, userID := principal(r)
	draft, err := h.service.RestoreRevision(r.Context(), workspaceID, userID, localizationID, revisionID, input.Version)
	if errors.Is(err, application.ErrConflict) {
		httpx.Error(w, 409, "draft_conflict", "draft changed before the revision was restored")
		return
	}
	if err != nil {
		httpx.Error(w, 422, "revision_restore_failed", err.Error())
		return
	}
	httpx.JSON(w, 200, draft)
}
func (h *HTTP) createPreview(w http.ResponseWriter, r *http.Request) {
	if !editorRequired(w, r) {
		return
	}
	localizationID, ok := parseID(w, r, "localizationID")
	if !ok {
		return
	}
	workspaceID, userID := principal(r)
	token, err := h.service.CreatePreview(r.Context(), workspaceID, userID, localizationID)
	if err != nil {
		httpx.Error(w, 422, "preview_failed", "could not create preview")
		return
	}
	httpx.JSON(w, 201, token)
}
func (h *HTTP) preview(w http.ResponseWriter, r *http.Request) {
	preview, err := h.service.GetPreview(r.Context(), chi.URLParam(r, "token"))
	if err != nil {
		httpx.Error(w, 404, "preview_not_found", "preview is invalid or expired")
		return
	}
	w.Header().Set("Cache-Control", "private, no-store")
	httpx.JSON(w, 200, preview)
}
