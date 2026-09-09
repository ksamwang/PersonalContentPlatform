package transport

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	contentdomain "github.com/ksamwang/PersonalContentPlatform/internal/content/domain"
	identity "github.com/ksamwang/PersonalContentPlatform/internal/identity/transport"
	"github.com/ksamwang/PersonalContentPlatform/internal/inbox/application"
	"github.com/ksamwang/PersonalContentPlatform/internal/inbox/domain"
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
		r.Route("/v1/workspaces/{workspaceID}/inbox", func(r chi.Router) {
			r.Get("/", h.list)
			r.Post("/", h.create)
			r.Post("/{itemID}:archive", h.archive)
			r.Post("/{itemID}:convert", h.convert)
			r.Post("/{itemID}:process", h.process)
		})
	})
}

func scope(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := identity.Principal(r.Context())
		workspaceID, err := uuid.Parse(chi.URLParam(r, "workspaceID"))
		if !ok || err != nil || p.WorkspaceID != workspaceID {
			httpx.Error(w, http.StatusForbidden, "workspace_forbidden", "workspace access denied")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func principal(r *http.Request) (uuid.UUID, uuid.UUID, string) {
	p, _ := identity.Principal(r.Context())
	return p.WorkspaceID, p.UserID, p.Role
}

func canWrite(w http.ResponseWriter, r *http.Request) bool {
	_, _, role := principal(r)
	if role != "owner" && role != "editor" {
		httpx.Error(w, http.StatusForbidden, "editor_required", "workspace editor access is required")
		return false
	}
	return true
}

func itemID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	value, err := uuid.Parse(chi.URLParam(r, "itemID"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_id", "invalid itemID")
		return uuid.Nil, false
	}
	return value, true
}

func (h *HTTP) list(w http.ResponseWriter, r *http.Request) {
	workspaceID, _, _ := principal(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := h.service.List(r.Context(), workspaceID, domain.State(r.URL.Query().Get("state")), limit)
	if err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, "inbox_invalid", err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *HTTP) create(w http.ResponseWriter, r *http.Request) {
	if !canWrite(w, r) {
		return
	}
	var input struct {
		Kind      domain.Kind `json:"kind"`
		RawText   string      `json:"raw_text"`
		SourceURL string      `json:"source_url"`
		AssetID   string      `json:"asset_id"`
	}
	if err := httpx.Decode(w, r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	workspaceID, userID, _ := principal(r)
	item, err := h.service.Create(r.Context(), workspaceID, userID, input.Kind, input.RawText, input.SourceURL, input.AssetID)
	if err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, "inbox_invalid", err.Error())
		return
	}
	httpx.JSON(w, http.StatusCreated, item)
}
func (h *HTTP) process(w http.ResponseWriter, r *http.Request) {
	if !canWrite(w, r) {
		return
	}
	id, ok := itemID(w, r)
	if !ok {
		return
	}
	ws, _, _ := principal(r)
	item, err := h.service.Process(r.Context(), ws, id)
	if err != nil {
		httpx.Error(w, 422, "inbox_processing_failed", err.Error())
		return
	}
	httpx.JSON(w, 200, item)
}

func (h *HTTP) archive(w http.ResponseWriter, r *http.Request) {
	if !canWrite(w, r) {
		return
	}
	id, ok := itemID(w, r)
	if !ok {
		return
	}
	workspaceID, _, _ := principal(r)
	if err := h.service.Archive(r.Context(), workspaceID, id); err != nil {
		status := http.StatusConflict
		if !errors.Is(err, application.ErrNotPending) {
			status = http.StatusInternalServerError
		}
		httpx.Error(w, status, "inbox_archive_failed", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *HTTP) convert(w http.ResponseWriter, r *http.Request) {
	if !canWrite(w, r) {
		return
	}
	id, ok := itemID(w, r)
	if !ok {
		return
	}
	var input struct {
		Type   contentdomain.Type `json:"type"`
		Locale string             `json:"locale"`
		Slug   string             `json:"slug"`
		Title  string             `json:"title"`
	}
	if err := httpx.Decode(w, r, &input); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return
	}
	workspaceID, userID, _ := principal(r)
	result, err := h.service.Convert(r.Context(), workspaceID, userID, id, input.Type, input.Locale, input.Slug, input.Title)
	if err != nil {
		status := http.StatusUnprocessableEntity
		if errors.Is(err, application.ErrNotPending) {
			status = http.StatusConflict
		}
		httpx.Error(w, status, "inbox_convert_failed", err.Error())
		return
	}
	httpx.JSON(w, http.StatusCreated, result)
}
