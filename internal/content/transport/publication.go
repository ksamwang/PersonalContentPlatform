package transport

import (
	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/httpx"
	"net/http"
)

type publicationInput struct {
	ContentID uuid.UUID `json:"content_id"`
	Locale    string    `json:"locale"`
	TargetID  uuid.UUID `json:"target_id"`
}

func (h *HTTP) publish(w http.ResponseWriter, r *http.Request) {
	if !editorRequired(w, r) {
		return
	}
	var input publicationInput
	if err := httpx.Decode(w, r, &input); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return
	}
	workspaceID, userID := principal(r)
	id, err := h.service.Publish(r.Context(), workspaceID, userID, input.ContentID, input.Locale, input.TargetID)
	if err != nil {
		httpx.Error(w, 422, "publication_rejected", err.Error())
		return
	}
	httpx.JSON(w, 202, map[string]any{"publication_id": id, "state": "queued"})
}
