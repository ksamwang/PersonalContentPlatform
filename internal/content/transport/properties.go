package transport

import (
	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/content/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/httpx"
	"net/http"
)

type propertiesInput struct {
	LocalizationID string            `json:"localization_id"`
	Slug           string            `json:"slug"`
	Visibility     domain.Visibility `json:"visibility"`
}

func (h *HTTP) updateProperties(w http.ResponseWriter, r *http.Request) {
	if !editorRequired(w, r) {
		return
	}
	contentID, ok := parseID(w, r, "contentID")
	if !ok {
		return
	}
	var input propertiesInput
	if err := httpx.Decode(w, r, &input); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return
	}
	localizationID, err := uuid.Parse(input.LocalizationID)
	if err != nil {
		httpx.Error(w, 400, "invalid_localization", "invalid localization_id")
		return
	}
	workspaceID, _ := principal(r)
	if err = h.service.UpdateProperties(r.Context(), workspaceID, contentID, localizationID, input.Slug, input.Visibility); err != nil {
		httpx.Error(w, 422, "properties_invalid", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *HTTP) readiness(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "localizationID")
	if !ok {
		return
	}
	workspaceID, _ := principal(r)
	result, err := h.service.Readiness(r.Context(), workspaceID, id)
	if err != nil {
		httpx.Error(w, 404, "localization_not_found", "localization was not found")
		return
	}
	httpx.JSON(w, 200, result)
}
