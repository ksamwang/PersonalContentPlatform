package transport

import (
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/asset/application"
	"github.com/ksamwang/PersonalContentPlatform/internal/asset/domain"
	identity "github.com/ksamwang/PersonalContentPlatform/internal/identity/transport"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/httpx"
	"io"
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
	r.Get("/v1/public/assets/{assetID}", h.publicContent)
	r.Get("/v1/public/assets/{assetID}/{recipe}", h.publicVariant)
	r.Group(func(r chi.Router) {
		r.Use(h.auth)
		r.Route("/v1/workspaces/{workspaceID}/assets", func(r chi.Router) {
			r.Use(scope)
			r.Get("/", h.list)
			r.Get("/{assetID}", h.get)
			r.Get("/{assetID}/usages", h.usages)
			r.Post("/{assetID}:archive", h.archive)
			r.Post("/{assetID}:replace", h.replace)
			r.Put("/uploads/{uploadID}", h.upload)
		})
		r.With(scope).Post("/v1/workspaces/{workspaceID}/assets:prepare-upload", h.prepare)
		r.With(scope).Post("/v1/workspaces/{workspaceID}/assets:finalize-upload", h.finalize)
	})
}
func assetID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	value, err := uuid.Parse(chi.URLParam(r, "assetID"))
	if err != nil {
		httpx.Error(w, 400, "invalid_asset", "invalid asset id")
		return uuid.Nil, false
	}
	return value, true
}
func (h *HTTP) get(w http.ResponseWriter, r *http.Request) {
	id, ok := assetID(w, r)
	if !ok {
		return
	}
	ws, _ := ids(r)
	item, err := h.service.Get(r.Context(), ws, id)
	if err != nil {
		httpx.Error(w, 404, "asset_not_found", "asset was not found")
		return
	}
	httpx.JSON(w, 200, item)
}
func (h *HTTP) usages(w http.ResponseWriter, r *http.Request) {
	id, ok := assetID(w, r)
	if !ok {
		return
	}
	ws, _ := ids(r)
	items, err := h.service.Usages(r.Context(), ws, id)
	if err != nil {
		httpx.Error(w, 500, "asset_usages_failed", "could not load asset usages")
		return
	}
	httpx.JSON(w, 200, map[string]any{"items": items})
}
func (h *HTTP) archive(w http.ResponseWriter, r *http.Request) {
	id, ok := assetID(w, r)
	if !ok {
		return
	}
	ws, _ := ids(r)
	if err := h.service.Archive(r.Context(), ws, id); err != nil {
		httpx.Error(w, 409, "asset_in_use", err.Error())
		return
	}
	w.WriteHeader(204)
}

type replaceInput struct {
	ReplacementAssetID uuid.UUID `json:"replacement_asset_id"`
}

func (h *HTTP) replace(w http.ResponseWriter, r *http.Request) {
	id, ok := assetID(w, r)
	if !ok {
		return
	}
	var in replaceInput
	if err := httpx.Decode(w, r, &in); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return
	}
	ws, _ := ids(r)
	item, err := h.service.Replace(r.Context(), ws, id, in.ReplacementAssetID)
	if err != nil {
		httpx.Error(w, 422, "asset_replace_failed", err.Error())
		return
	}
	httpx.JSON(w, 200, item)
}
func (h *HTTP) publicContent(w http.ResponseWriter, r *http.Request) {
	assetID, err := uuid.Parse(chi.URLParam(r, "assetID"))
	if err != nil {
		httpx.Error(w, 404, "asset_not_found", "asset was not found")
		return
	}
	reader, object, err := h.service.OpenPublic(r.Context(), assetID)
	if err != nil {
		httpx.Error(w, 404, "asset_not_found", "asset was not found")
		return
	}
	serveObject(w, reader, object)
}
func (h *HTTP) publicVariant(w http.ResponseWriter, r *http.Request) {
	assetID, err := uuid.Parse(chi.URLParam(r, "assetID"))
	if err != nil {
		httpx.Error(w, 404, "asset_not_found", "asset was not found")
		return
	}
	reader, object, err := h.service.OpenPublicVariant(r.Context(), assetID, chi.URLParam(r, "recipe"))
	if err != nil {
		httpx.Error(w, 404, "asset_not_found", "asset was not found")
		return
	}
	serveObject(w, reader, object)
}
func serveObject(w http.ResponseWriter, reader io.ReadCloser, object domain.StoredObject) {
	defer reader.Close()
	w.Header().Set("Content-Type", object.MIME)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("Content-Length", strconv.FormatInt(object.Size, 10))
	_, _ = io.Copy(w, reader)
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
func ids(r *http.Request) (uuid.UUID, uuid.UUID) {
	p, _ := identity.Principal(r.Context())
	return p.WorkspaceID, p.UserID
}

type prepareInput struct {
	Filename string `json:"filename"`
	MIME     string `json:"mime"`
	Size     int64  `json:"size"`
}

func (h *HTTP) prepare(w http.ResponseWriter, r *http.Request) {
	var in prepareInput
	if err := httpx.Decode(w, r, &in); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return
	}
	ws, user := ids(r)
	plan, err := h.service.Prepare(r.Context(), ws, user, in.Filename, in.MIME, in.Size)
	if err != nil {
		httpx.Error(w, 422, "upload_rejected", err.Error())
		return
	}
	httpx.JSON(w, 201, plan)
}
func (h *HTTP) upload(w http.ResponseWriter, r *http.Request) {
	uploadID, err := uuid.Parse(chi.URLParam(r, "uploadID"))
	if err != nil {
		httpx.Error(w, 400, "invalid_upload", "invalid upload id")
		return
	}
	ws, _ := ids(r)
	r.Body = http.MaxBytesReader(w, r.Body, application.MaxUploadSize)
	if err = h.service.Upload(r.Context(), ws, uploadID, r.Body, r.ContentLength); err != nil {
		httpx.Error(w, 422, "upload_failed", err.Error())
		return
	}
	w.WriteHeader(204)
}

type finalizeInput struct {
	UploadID uuid.UUID `json:"upload_id"`
}

func (h *HTTP) finalize(w http.ResponseWriter, r *http.Request) {
	var in finalizeInput
	if err := httpx.Decode(w, r, &in); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return
	}
	ws, user := ids(r)
	asset, err := h.service.Complete(r.Context(), ws, user, in.UploadID)
	if err != nil {
		httpx.Error(w, 422, "finalize_failed", err.Error())
		return
	}
	httpx.JSON(w, 201, asset)
}
func (h *HTTP) list(w http.ResponseWriter, r *http.Request) {
	ws, _ := ids(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := h.service.List(r.Context(), ws, limit)
	if err != nil {
		httpx.Error(w, 500, "assets_failed", "could not load assets")
		return
	}
	httpx.JSON(w, 200, map[string]any{"items": items})
}
