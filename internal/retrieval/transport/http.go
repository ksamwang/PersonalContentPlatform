package transport

import (
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	identity "github.com/ksamwang/PersonalContentPlatform/internal/identity/transport"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/httpx"
	"github.com/ksamwang/PersonalContentPlatform/internal/retrieval/application"
	"net/http"
	"strconv"
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
		r.Post("/v1/workspaces/{workspaceID}/search:index", h.index)
		r.Get("/v1/workspaces/{workspaceID}/hybrid-search", h.search)
		r.Post("/v1/workspaces/{workspaceID}/rag", h.rag)
	})
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
func ws(r *http.Request) uuid.UUID { p, _ := identity.Principal(r.Context()); return p.WorkspaceID }
func (h *HTTP) index(w http.ResponseWriter, r *http.Request) {
	count, err := h.service.Index(r.Context(), ws(r))
	if err != nil {
		httpx.Error(w, 422, "search_index_failed", err.Error())
		return
	}
	httpx.JSON(w, 200, map[string]any{"chunks": count})
}
func (h *HTTP) search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		httpx.Error(w, 400, "query_required", "query is required")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := h.service.Search(r.Context(), ws(r), q, r.URL.Query().Get("locale"), limit)
	if err != nil {
		httpx.Error(w, 422, "hybrid_search_failed", err.Error())
		return
	}
	httpx.JSON(w, 200, map[string]any{"items": items})
}
func (h *HTTP) rag(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Query  string `json:"query"`
		Locale string `json:"locale"`
	}
	if err := httpx.Decode(w, r, &input); err != nil || input.Query == "" {
		httpx.Error(w, 400, "invalid_request", "query is required")
		return
	}
	answer, sources, err := h.service.RAG(r.Context(), ws(r), input.Query, input.Locale)
	if err != nil {
		httpx.Error(w, 422, "rag_failed", err.Error())
		return
	}
	httpx.JSON(w, 200, map[string]any{"answer": answer, "sources": sources})
}
