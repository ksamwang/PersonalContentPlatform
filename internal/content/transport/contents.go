package transport

import (
	"github.com/ksamwang/PersonalContentPlatform/internal/content/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/httpx"
	"net/http"
	"strconv"
)

type createLocalizationInput struct {
	Locale       string `json:"locale"`
	Slug         string `json:"slug"`
	SourceLocale string `json:"source_locale"`
}

func (h *HTTP) createLocalization(w http.ResponseWriter, r *http.Request) {
	if !editorRequired(w, r) {
		return
	}
	contentID, ok := parseID(w, r, "contentID")
	if !ok {
		return
	}
	var input createLocalizationInput
	if err := httpx.Decode(w, r, &input); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return
	}
	workspaceID, userID := principal(r)
	content, err := h.service.CreateLocalization(r.Context(), workspaceID, userID, contentID, input.Locale, input.Slug, input.SourceLocale)
	if err != nil {
		httpx.Error(w, 422, "localization_invalid", err.Error())
		return
	}
	httpx.JSON(w, 201, content)
}

type createInput struct {
	Type   domain.Type `json:"type"`
	Locale string      `json:"locale"`
	Slug   string      `json:"slug"`
	Title  string      `json:"title"`
}

func (h *HTTP) create(w http.ResponseWriter, r *http.Request) {
	if !editorRequired(w, r) {
		return
	}
	var input createInput
	if err := httpx.Decode(w, r, &input); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return
	}
	workspaceID, userID := principal(r)
	content, err := h.service.Create(r.Context(), workspaceID, userID, input.Type, input.Locale, input.Slug, input.Title)
	if err != nil {
		httpx.Error(w, 422, "content_invalid", err.Error())
		return
	}
	httpx.JSON(w, 201, content)
}
func (h *HTTP) list(w http.ResponseWriter, r *http.Request) {
	workspaceID, _ := principal(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := h.service.List(r.Context(), workspaceID, domain.ListFilter{
		Locale: r.URL.Query().Get("locale"), State: r.URL.Query().Get("state"),
		Type: domain.Type(r.URL.Query().Get("type")), Visibility: domain.Visibility(r.URL.Query().Get("visibility")),
		Tag: r.URL.Query().Get("tag"), Limit: limit,
	})
	if err != nil {
		httpx.Error(w, 500, "content_list_failed", "could not load content")
		return
	}
	httpx.JSON(w, 200, map[string]any{"items": items})
}

func (h *HTTP) archive(w http.ResponseWriter, r *http.Request) {
	if !editorRequired(w, r) {
		return
	}
	id, ok := parseID(w, r, "contentID")
	if !ok {
		return
	}
	workspaceID, _ := principal(r)
	if err := h.service.Archive(r.Context(), workspaceID, id); err != nil {
		httpx.Error(w, 404, "content_not_found", "content was not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *HTTP) restore(w http.ResponseWriter, r *http.Request) {
	if !editorRequired(w, r) {
		return
	}
	id, ok := parseID(w, r, "contentID")
	if !ok {
		return
	}
	workspaceID, _ := principal(r)
	if err := h.service.Restore(r.Context(), workspaceID, id); err != nil {
		httpx.Error(w, 404, "content_not_found", "content was not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *HTTP) remove(w http.ResponseWriter, r *http.Request) {
	if !editorRequired(w, r) {
		return
	}
	id, ok := parseID(w, r, "contentID")
	if !ok {
		return
	}
	workspaceID, _ := principal(r)
	if err := h.service.SoftDelete(r.Context(), workspaceID, id); err != nil {
		httpx.Error(w, 404, "content_not_found", "content was not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *HTTP) get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "contentID")
	if !ok {
		return
	}
	workspaceID, _ := principal(r)
	item, err := h.service.Get(r.Context(), workspaceID, id)
	if err != nil {
		httpx.Error(w, 404, "content_not_found", "content was not found")
		return
	}
	httpx.JSON(w, 200, item)
}
func (h *HTTP) search(w http.ResponseWriter, r *http.Request) {
	workspaceID, _ := principal(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := h.service.Search(r.Context(), workspaceID, r.URL.Query().Get("locale"), r.URL.Query().Get("q"), limit)
	if err != nil {
		httpx.Error(w, 500, "search_failed", "search is temporarily unavailable")
		return
	}
	httpx.JSON(w, 200, map[string]any{"items": items})
}
