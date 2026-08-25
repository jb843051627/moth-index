package handler

import (
	"net/http"

	"github.com/jb843051627/moth-index/internal/model"
)

func (h *Handler) listStations(w http.ResponseWriter, r *http.Request) {
	page := model.ParsePage(intParam(r, "limit", 25), intParam(r, "offset", 0))
	includeRetired := r.URL.Query().Get("include_retired") == "1"
	values, err := h.services.Stations.List(r.Context(), page, includeRetired)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, values)
}

func (h *Handler) createStation(w http.ResponseWriter, r *http.Request) {
	var input model.Station
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, err)
		return
	}
	station, err := h.services.Stations.Create(r.Context(), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, station)
}

func (h *Handler) getStation(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		writeError(w, err)
		return
	}
	station, err := h.services.Stations.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, station)
}

func (h *Handler) retireStation(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err == nil {
		err = h.services.Stations.Retire(r.Context(), id)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) reactivateStation(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err == nil {
		err = h.services.Stations.Reactivate(r.Context(), id)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
