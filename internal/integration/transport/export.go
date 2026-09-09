package transport

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/httpx"
)

func (h *HTTP) exportArchive(w http.ResponseWriter, r *http.Request) {
	if h.archives == nil {
		httpx.Error(w, 501, "export_unavailable", "archive export is unavailable")
		return
	}
	caller := principal(r)
	body, name, mimeType, err := h.archives.Build(r.Context(), caller.WorkspaceID, caller.UserID, chi.URLParam(r, "format"))
	if err != nil {
		httpx.Error(w, 422, "export_failed", err.Error())
		return
	}
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	w.Header().Set("Content-Type", mimeType)
	_, _ = w.Write(body)
}

func (h *HTTP) importContent(w http.ResponseWriter, r *http.Request) {
	caller := principal(r)
	if caller.Role != "owner" && caller.Role != "editor" {
		httpx.Error(w, 403, "editor_required", "workspace editor access is required")
		return
	}
	if h.imports == nil {
		httpx.Error(w, 501, "import_unavailable", "import is unavailable")
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 250<<20))
	if err != nil {
		httpx.Error(w, 400, "import_invalid", err.Error())
		return
	}
	format := chi.URLParam(r, "format")
	var report any
	if r.URL.Query().Get("mode") == "apply" {
		report, err = h.imports.Apply(r.Context(), caller.WorkspaceID, caller.UserID, format, body)
	} else {
		report, err = h.imports.Preview(r.Context(), caller.WorkspaceID, format, body)
	}
	if err != nil {
		httpx.Error(w, 422, "import_invalid", err.Error())
		return
	}
	httpx.JSON(w, 200, report)
}

func (h *HTTP) exportManifest(w http.ResponseWriter, r *http.Request) {
	caller := principal(r)
	manifest, err := h.exports.Build(r.Context(), caller.WorkspaceID, caller.UserID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "export_failed", "workspace export failed")
		return
	}
	w.Header().Set("Content-Disposition", `attachment; filename="workspace-manifest.json"`)
	httpx.JSON(w, http.StatusOK, manifest)
}
