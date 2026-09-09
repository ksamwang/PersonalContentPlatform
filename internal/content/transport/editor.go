package transport

import (
	"encoding/json"
	"errors"
	"github.com/ksamwang/PersonalContentPlatform/internal/content/application"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/httpx"
	"net/http"
	"strconv"
)

type draftInput struct {
	Version  int             `json:"version"`
	Title    string          `json:"title"`
	Summary  string          `json:"summary"`
	Body     json.RawMessage `json:"body"`
	Metadata json.RawMessage `json:"metadata"`
}

func (h *HTTP) getDraft(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "localizationID")
	if !ok {
		return
	}
	workspaceID, _ := principal(r)
	draft, err := h.service.GetDraft(r.Context(), workspaceID, id)
	if err != nil {
		httpx.Error(w, 404, "draft_not_found", "draft was not found")
		return
	}
	httpx.JSON(w, 200, draft)
}

func (h *HTTP) saveDraft(w http.ResponseWriter, r *http.Request) {
	if !editorRequired(w, r) {
		return
	}
	id, ok := parseID(w, r, "localizationID")
	if !ok {
		return
	}
	var input draftInput
	if err := httpx.Decode(w, r, &input); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return
	}
	workspaceID, userID := principal(r)
	draft, err := h.service.SaveDraft(r.Context(), workspaceID, userID, id, input.Version, input.Title, input.Summary, input.Body, input.Metadata)
	if errors.Is(err, application.ErrConflict) {
		httpx.Error(w, 409, "draft_conflict", "draft changed in another session; reload before saving")
		return
	}
	if err != nil {
		httpx.Error(w, 422, "draft_invalid", err.Error())
		return
	}
	w.Header().Set("ETag", `"`+strconv.Itoa(draft.Version)+`"`)
	httpx.JSON(w, 200, draft)
}

type sealInput struct {
	Version int `json:"version"`
}

func (h *HTTP) sealRevision(w http.ResponseWriter, r *http.Request) {
	if !editorRequired(w, r) {
		return
	}
	id, ok := parseID(w, r, "localizationID")
	if !ok {
		return
	}
	var input sealInput
	if err := httpx.Decode(w, r, &input); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return
	}
	workspaceID, userID := principal(r)
	revision, err := h.service.Seal(r.Context(), workspaceID, userID, id, input.Version)
	if errors.Is(err, application.ErrConflict) {
		httpx.Error(w, 409, "draft_conflict", "draft changed before revision was created")
		return
	}
	if err != nil {
		httpx.Error(w, 422, "revision_failed", err.Error())
		return
	}
	httpx.JSON(w, 201, revision)
}
func (h *HTTP) markReady(w http.ResponseWriter, r *http.Request) {
	if !editorRequired(w, r) {
		return
	}
	id, ok := parseID(w, r, "localizationID")
	if !ok {
		return
	}
	workspaceID, _ := principal(r)
	if err := h.service.MarkReady(r.Context(), workspaceID, id); err != nil {
		httpx.Error(w, 422, "not_ready", "localization cannot be marked ready")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
