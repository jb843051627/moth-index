package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/jb843051627/moth-index/internal/model"
)

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	status := model.Health{Status: "ok", Database: "ok", QueueSize: h.queue.Size(), Now: time.Now().UTC().Format(time.RFC3339Nano)}
	if err := h.services.DB.Ping(ctx); err != nil {
		status.Status = "degraded"
		status.Database = err.Error()
		writeJSON(w, http.StatusServiceUnavailable, status)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (h *Handler) homeData(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"service": "moth-index"})
}
