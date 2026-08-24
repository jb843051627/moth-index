package handler

import "net/http"

func (h *Handler) analyzeBatch(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		writeError(w, err)
		return
	}
	result, err := h.services.Analysis.Analyze(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) batchFindings(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		writeError(w, err)
		return
	}
	findings, err := h.services.Analysis.Findings(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, findings)
}

func (h *Handler) phenologyReport(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		writeError(w, err)
		return
	}
	result, err := h.services.Phenology.Report(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) seasonScore(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		writeError(w, err)
		return
	}
	score, err := h.services.Seasonal.Score(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"batch_id": id, "score": score})
}

func (h *Handler) maintenanceEvents(w http.ResponseWriter, r *http.Request) {
	entity := r.URL.Query().Get("entity")
	entityID := int64Param(r, "entity_id")
	events, err := h.services.Maintenance.Events(r.Context(), entity, entityID, intParam(r, "limit", 50))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (h *Handler) queueMetrics(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.queue.Metrics())
}
