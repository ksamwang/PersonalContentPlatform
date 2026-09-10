package transport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	identity "github.com/ksamwang/PersonalContentPlatform/internal/identity/transport"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/httpx"
	"github.com/ksamwang/PersonalContentPlatform/internal/task/application"
)

type HTTP struct {
	queue *application.Queue
	auth  func(http.Handler) http.Handler
}

func NewHTTP(queue *application.Queue, auth func(http.Handler) http.Handler) *HTTP {
	return &HTTP{queue: queue, auth: auth}
}

func (h *HTTP) Register(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(h.auth)
		r.Get("/v1/workspaces/{workspaceID}/tasks/{taskID}", h.status)
	})
}

func (h *HTTP) status(w http.ResponseWriter, r *http.Request) {
	p, ok := identity.Principal(r.Context())
	ws, wsErr := uuid.Parse(chi.URLParam(r, "workspaceID"))
	jobID, idErr := uuid.Parse(chi.URLParam(r, "taskID"))
	if !ok || wsErr != nil || idErr != nil || p.WorkspaceID != ws {
		httpx.Error(w, http.StatusForbidden, "workspace_forbidden", "workspace access denied")
		return
	}
	value, err := h.queue.Status(r.Context(), ws, jobID)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "task_not_found", "background task was not found")
		return
	}
	httpx.JSON(w, http.StatusOK, value)
}
