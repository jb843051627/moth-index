package handler

import (
	"net/http"
)

func (h *Handler) batchSummary(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		writeError(w, err)
		return
	}
	summary, err := h.services.Reports.Summary(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (h *Handler) evaluateBatch(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		writeError(w, err)
		return
	}
	summary, err := h.services.Quality.EvaluateBatch(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (h *Handler) dailyReport(w http.ResponseWriter, r *http.Request) {
	values, err := h.services.Reports.Daily(r.Context(), r.URL.Query().Get("from"), r.URL.Query().Get("to"), intParam(r, "limit", 31))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, values)
}

func (h *Handler) dailyCSV(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	if err := h.services.Reports.WriteDailyCSV(r.Context(), w, r.URL.Query().Get("from"), r.URL.Query().Get("to"), intParam(r, "limit", 31)); err != nil {
		writeError(w, err)
	}
}

func (h *Handler) timeline(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		writeError(w, err)
		return
	}
	values, err := h.services.Reports.Timeline(r.Context(), id, intParam(r, "limit", 200))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, values)
}
