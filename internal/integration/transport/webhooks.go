package transport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/httpx"
)

type createWebhookInput struct {
	Name       string   `json:"name"`
	URL        string   `json:"url"`
	SecretRef  string   `json:"secret_ref"`
	Secret     string   `json:"secret"`
	EventTypes []string `json:"event_types"`
}

func (h *HTTP) createWebhook(w http.ResponseWriter, r *http.Request) {
	var input createWebhookInput
	if err := httpx.Decode(w, r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	endpoint, err := h.webhooks.Create(r.Context(), principal(r).WorkspaceID, input.Name, input.URL, input.SecretRef, input.Secret, input.EventTypes)
	if err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, "webhook_invalid", err.Error())
		return
	}
	httpx.JSON(w, http.StatusCreated, endpoint)
}

func (h *HTTP) listWebhooks(w http.ResponseWriter, r *http.Request) {
	items, err := h.webhooks.List(r.Context(), principal(r).WorkspaceID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "webhooks_failed", "could not load webhooks")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

type setWebhookEnabledInput struct {
	Enabled bool `json:"enabled"`
}

func (h *HTTP) setWebhookEnabled(w http.ResponseWriter, r *http.Request) {
	endpointID, err := uuid.Parse(chi.URLParam(r, "endpointID"))
	var input setWebhookEnabledInput
	if err != nil || httpx.Decode(w, r, &input) != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", "invalid webhook update")
		return
	}
	if err = h.webhooks.SetEnabled(r.Context(), principal(r).WorkspaceID, endpointID, input.Enabled); err != nil {
		httpx.Error(w, http.StatusNotFound, "webhook_not_found", "webhook endpoint was not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *HTTP) deleteWebhook(w http.ResponseWriter, r *http.Request) {
	endpointID, err := uuid.Parse(chi.URLParam(r, "endpointID"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_id", "invalid webhook endpoint id")
		return
	}
	if err = h.webhooks.Delete(r.Context(), principal(r).WorkspaceID, endpointID); err != nil {
		httpx.Error(w, http.StatusNotFound, "webhook_not_found", "webhook endpoint was not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
