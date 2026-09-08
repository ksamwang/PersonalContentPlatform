package transport

import (
	"net/http"

	"github.com/ksamwang/PersonalContentPlatform/internal/platform/httpx"
)

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
