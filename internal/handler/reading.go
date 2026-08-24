package handler

import (
	"net/http"

	"github.com/jb843051627/moth-index/internal/model"
)

func (h *Handler) listReadings(w http.ResponseWriter, r *http.Request) {
	batchID, err := idParam(r)
	if err != nil {
		writeError(w, err)
		return
	}
	values, err := h.services.Readings.List(r.Context(), batchID, r.URL.Query().Get("from"), r.URL.Query().Get("to"), intParam(r, "limit", 200))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, values)
}

func (h *Handler) ingestReading(w http.ResponseWriter, r *http.Request) {
	batchID, err := idParam(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var input model.Reading
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, err)
		return
	}
	input.BatchID = batchID
	reading, err := h.services.Readings.Ingest(r.Context(), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, reading)
}
