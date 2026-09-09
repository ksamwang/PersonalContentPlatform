package transport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/collection/application"
	identity "github.com/ksamwang/PersonalContentPlatform/internal/identity/transport"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/httpx"
)

type HTTP struct {
	service *application.Service
	auth    func(http.Handler) http.Handler
}

func NewHTTP(service *application.Service, auth func(http.Handler) http.Handler) *HTTP {
	return &HTTP{service: service, auth: auth}
}
func (h *HTTP) Register(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(h.auth, scope)
		r.Get("/v1/workspaces/{workspaceID}/collections", h.list)
		r.Post("/v1/workspaces/{workspaceID}/collections", h.create)
		r.Get("/v1/workspaces/{workspaceID}/collections/{collectionID}", h.get)
		r.Put("/v1/workspaces/{workspaceID}/collections/{collectionID}", h.update)
		r.Post("/v1/workspaces/{workspaceID}/collections/{collectionID}/sections", h.addSection)
		r.Post("/v1/workspaces/{workspaceID}/collections/{collectionID}/sections/{sectionID}:move", h.moveSection)
		r.Post("/v1/workspaces/{workspaceID}/collection-sections/{sectionID}/items", h.addItem)
		r.Post("/v1/workspaces/{workspaceID}/collection-items/{itemID}:move", h.moveItem)
		r.Delete("/v1/workspaces/{workspaceID}/collection-items/{itemID}", h.removeItem)
	})
}
func scope(next http.Handler) http.Handler {
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
func principal(r *http.Request) (uuid.UUID, string) {
	p, _ := identity.Principal(r.Context())
	return p.WorkspaceID, p.Role
}
func writable(w http.ResponseWriter, r *http.Request) bool {
	_, role := principal(r)
	if role != "owner" && role != "editor" {
		httpx.Error(w, 403, "editor_required", "workspace editor access is required")
		return false
	}
	return true
}
func parse(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	v, err := uuid.Parse(chi.URLParam(r, name))
	if err != nil {
		httpx.Error(w, 400, "invalid_id", "invalid "+name)
		return uuid.Nil, false
	}
	return v, true
}

type collectionInput struct {
	Title      string `json:"title"`
	Slug       string `json:"slug"`
	Visibility string `json:"visibility"`
}

func (h *HTTP) list(w http.ResponseWriter, r *http.Request) {
	ws, _ := principal(r)
	v, err := h.service.List(r.Context(), ws)
	if err != nil {
		httpx.Error(w, 500, "collections_failed", "could not load collections")
		return
	}
	httpx.JSON(w, 200, map[string]any{"items": v})
}
func (h *HTTP) create(w http.ResponseWriter, r *http.Request) {
	if !writable(w, r) {
		return
	}
	var in collectionInput
	if err := httpx.Decode(w, r, &in); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return
	}
	ws, _ := principal(r)
	v, err := h.service.Create(r.Context(), ws, in.Title, in.Slug, in.Visibility)
	if err != nil {
		httpx.Error(w, 422, "collection_invalid", err.Error())
		return
	}
	httpx.JSON(w, 201, v)
}
func (h *HTTP) get(w http.ResponseWriter, r *http.Request) {
	id, ok := parse(w, r, "collectionID")
	if !ok {
		return
	}
	ws, _ := principal(r)
	v, err := h.service.Get(r.Context(), ws, id)
	if err != nil {
		httpx.Error(w, 404, "collection_not_found", "collection was not found")
		return
	}
	httpx.JSON(w, 200, v)
}
func (h *HTTP) update(w http.ResponseWriter, r *http.Request) {
	if !writable(w, r) {
		return
	}
	id, ok := parse(w, r, "collectionID")
	if !ok {
		return
	}
	var in collectionInput
	if err := httpx.Decode(w, r, &in); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return
	}
	ws, _ := principal(r)
	v, err := h.service.Update(r.Context(), ws, id, in.Title, in.Slug, in.Visibility)
	if err != nil {
		httpx.Error(w, 422, "collection_invalid", err.Error())
		return
	}
	httpx.JSON(w, 200, v)
}
func (h *HTTP) addSection(w http.ResponseWriter, r *http.Request) {
	if !writable(w, r) {
		return
	}
	id, ok := parse(w, r, "collectionID")
	if !ok {
		return
	}
	var in struct {
		Title string `json:"title"`
	}
	if err := httpx.Decode(w, r, &in); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return
	}
	ws, _ := principal(r)
	v, err := h.service.AddSection(r.Context(), ws, id, in.Title)
	if err != nil {
		httpx.Error(w, 422, "section_invalid", err.Error())
		return
	}
	httpx.JSON(w, 201, v)
}
func direction(w http.ResponseWriter, r *http.Request) (int, bool) {
	var in struct {
		Direction int `json:"direction"`
	}
	if err := httpx.Decode(w, r, &in); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return 0, false
	}
	return in.Direction, true
}
func (h *HTTP) moveSection(w http.ResponseWriter, r *http.Request) {
	if !writable(w, r) {
		return
	}
	cid, ok := parse(w, r, "collectionID")
	if !ok {
		return
	}
	sid, ok := parse(w, r, "sectionID")
	if !ok {
		return
	}
	d, ok := direction(w, r)
	if !ok {
		return
	}
	ws, _ := principal(r)
	if err := h.service.MoveSection(r.Context(), ws, cid, sid, d); err != nil {
		httpx.Error(w, 422, "section_move_failed", err.Error())
		return
	}
	w.WriteHeader(204)
}
func (h *HTTP) addItem(w http.ResponseWriter, r *http.Request) {
	if !writable(w, r) {
		return
	}
	sid, ok := parse(w, r, "sectionID")
	if !ok {
		return
	}
	var in struct {
		ObjectID   uuid.UUID `json:"object_id"`
		Annotation string    `json:"annotation"`
	}
	if err := httpx.Decode(w, r, &in); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return
	}
	ws, _ := principal(r)
	v, err := h.service.AddItem(r.Context(), ws, sid, in.ObjectID, in.Annotation)
	if err != nil {
		httpx.Error(w, 422, "collection_item_invalid", err.Error())
		return
	}
	httpx.JSON(w, 201, v)
}
func (h *HTTP) moveItem(w http.ResponseWriter, r *http.Request) {
	if !writable(w, r) {
		return
	}
	id, ok := parse(w, r, "itemID")
	if !ok {
		return
	}
	d, ok := direction(w, r)
	if !ok {
		return
	}
	ws, _ := principal(r)
	if err := h.service.MoveItem(r.Context(), ws, id, d); err != nil {
		httpx.Error(w, 422, "collection_item_move_failed", err.Error())
		return
	}
	w.WriteHeader(204)
}
func (h *HTTP) removeItem(w http.ResponseWriter, r *http.Request) {
	if !writable(w, r) {
		return
	}
	id, ok := parse(w, r, "itemID")
	if !ok {
		return
	}
	ws, _ := principal(r)
	if err := h.service.RemoveItem(r.Context(), ws, id); err != nil {
		httpx.Error(w, 404, "collection_item_not_found", "collection item was not found")
		return
	}
	w.WriteHeader(204)
}
