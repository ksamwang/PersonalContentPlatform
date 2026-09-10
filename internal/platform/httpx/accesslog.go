package httpx

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type responseStatus struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *responseStatus) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseStatus) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(body)
	w.bytes += n
	return n, err
}

func (w *responseStatus) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		writer := &responseStatus{ResponseWriter: w}
		next.ServeHTTP(writer, r)
		status := writer.status
		if status == 0 {
			status = http.StatusOK
		}
		slog.Info("http request",
			"request_id", middleware.GetReqID(r.Context()),
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"bytes", writer.bytes,
			"duration_ms", time.Since(started).Milliseconds(),
		)
	})
}
