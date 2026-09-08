package transport

import (
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	identity "github.com/ksamwang/PersonalContentPlatform/internal/identity/transport"
	"github.com/ksamwang/PersonalContentPlatform/internal/knowledge/application"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/httpx"
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
		r.Use(h.auth)
		r.Route("/v1/workspaces/{workspaceID}/knowledge", func(r chi.Router) {
			r.Use(scope)
			r.Get("/entities", h.list)
			r.Post("/entities", h.create)
			r.Post("/entities/{entityID}/aliases", h.alias)
			r.Post("/relations", h.relate)
			r.Post("/relations/{relationID}:confirm", h.confirm)
			r.Get("/objects/{objectID}/relations", h.relations)
		})
	})
}
func scope(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, _ := identity.Principal(r.Context())
		ws, err := uuid.Parse(chi.URLParam(r, "workspaceID"))
		if err != nil || p.WorkspaceID != ws {
			httpx.Error(w, 403, "workspace_forbidden", "workspace access denied")
			return
		}
		next.ServeHTTP(w, r)
	})
}
func ws(r *http.Request) uuid.UUID { p, _ := identity.Principal(r.Context()); return p.WorkspaceID }

type entityInput struct {
	Type        string `json:"type"`
	Name        string `json:"canonical_name"`
	Description string `json:"description"`
}

func (h *HTTP) create(w http.ResponseWriter, r *http.Request) {
	var in entityInput
	if err := httpx.Decode(w, r, &in); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return
	}
	v, err := h.service.CreateEntity(r.Context(), ws(r), in.Type, in.Name, in.Description)
	if err != nil {
		httpx.Error(w, 422, "entity_invalid", err.Error())
		return
	}
	httpx.JSON(w, 201, v)
}
func (h *HTTP) list(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := h.service.ListEntities(r.Context(), ws(r), r.URL.Query().Get("q"), limit)
	if err != nil {
		httpx.Error(w, 500, "entities_failed", "could not load entities")
		return
	}
	httpx.JSON(w, 200, map[string]any{"items": items})
}

type aliasInput struct{ Value, Locale string }

func (h *HTTP) alias(w http.ResponseWriter, r *http.Request) {
	idValue, err := uuid.Parse(chi.URLParam(r, "entityID"))
	var in aliasInput
	if err != nil || httpx.Decode(w, r, &in) != nil {
		httpx.Error(w, 400, "invalid_request", "invalid alias request")
		return
	}
	if err = h.service.AddAlias(r.Context(), ws(r), idValue, in.Value, in.Locale); err != nil {
		httpx.Error(w, 422, "alias_invalid", err.Error())
		return
	}
	w.WriteHeader(204)
}

type relationInput struct {
	SourceID  uuid.UUID `json:"source_id"`
	TargetID  uuid.UUID `json:"target_id"`
	Predicate string    `json:"predicate"`
	Confirmed bool      `json:"confirmed"`
}

func (h *HTTP) relate(w http.ResponseWriter, r *http.Request) {
	var in relationInput
	if err := httpx.Decode(w, r, &in); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return
	}
	v, err := h.service.CreateRelation(r.Context(), ws(r), in.SourceID, in.Predicate, in.TargetID, in.Confirmed)
	if err != nil {
		httpx.Error(w, 422, "relation_invalid", err.Error())
		return
	}
	httpx.JSON(w, 201, v)
}
func (h *HTTP) confirm(w http.ResponseWriter, r *http.Request) {
	idValue, err := uuid.Parse(chi.URLParam(r, "relationID"))
	if err != nil {
		httpx.Error(w, 400, "invalid_id", "invalid relation id")
		return
	}
	if err = h.service.Confirm(r.Context(), ws(r), idValue); err != nil {
		httpx.Error(w, 500, "confirm_failed", "could not confirm relation")
		return
	}
	w.WriteHeader(204)
}
func (h *HTTP) relations(w http.ResponseWriter, r *http.Request) {
	idValue, err := uuid.Parse(chi.URLParam(r, "objectID"))
	if err != nil {
		httpx.Error(w, 400, "invalid_id", "invalid object id")
		return
	}
	items, err := h.service.Relations(r.Context(), ws(r), idValue)
	if err != nil {
		httpx.Error(w, 500, "relations_failed", "could not load relations")
		return
	}
	httpx.JSON(w, 200, map[string]any{"items": items})
}
