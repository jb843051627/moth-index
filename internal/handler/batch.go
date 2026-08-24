package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jb843051627/moth-index/internal/model"
)

func (h *Handler) listBatches(w http.ResponseWriter, r *http.Request) {
	stationID := int64Param(r, "station_id")
	filter := model.BatchFilter{StationID: stationID, Status: r.URL.Query().Get("status"), From: r.URL.Query().Get("from"), To: r.URL.Query().Get("to")}
	values, err := h.services.Batches.List(r.Context(), filter, model.ParsePage(intParam(r, "limit", 25), intParam(r, "offset", 0)))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, values)
}

func (h *Handler) openBatch(w http.ResponseWriter, r *http.Request) {
	var input model.NightBatch
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, err)
		return
	}
	batch, err := h.services.Batches.Open(r.Context(), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, batch)
}

func (h *Handler) getBatch(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		writeError(w, err)
		return
	}
	batch, err := h.services.Batches.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, batch)
}

func (h *Handler) reviewBatch(w http.ResponseWriter, r *http.Request) {
	h.changeBatchStatus(w, r, h.services.Batches.StartReview)
}

func (h *Handler) closeBatch(w http.ResponseWriter, r *http.Request) {
	h.changeBatchStatus(w, r, h.services.Batches.Close)
}

func (h *Handler) archiveBatch(w http.ResponseWriter, r *http.Request) {
	h.changeBatchStatus(w, r, h.services.Archive.Archive)
}

func (h *Handler) changeBatchStatus(w http.ResponseWriter, r *http.Request, change func(context.Context, int64) error) {
	id, err := idParam(r)
	if err == nil {
		err = change(r.Context(), id)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func int64Param(r *http.Request, name string) int64 {
	value := r.URL.Query().Get(name)
	if value == "" {
		return 0
	}
	var number int64
	_, _ = fmt.Sscan(value, &number)
	return number
}
