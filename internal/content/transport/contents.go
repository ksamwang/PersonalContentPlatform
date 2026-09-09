package transport

import (
	"github.com/ksamwang/PersonalContentPlatform/internal/content/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/httpx"
	"net/http"
	"strconv"
)

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
	items, err := h.service.List(r.Context(), workspaceID, r.URL.Query().Get("locale"), r.URL.Query().Get("state"), limit)
	if err != nil {
		httpx.Error(w, 500, "content_list_failed", "could not load content")
		return
	}
	httpx.JSON(w, 200, map[string]any{"items": items})
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
