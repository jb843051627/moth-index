package handler

import (
	"context"
	"net/http"

	"github.com/jb843051627/moth-index/internal/model"
)

func (h *Handler) listTraps(w http.ResponseWriter, r *http.Request) {
	stationID, err := idParam(r)
	if err != nil {
		writeError(w, err)
		return
	}
	values, err := h.services.Traps.List(r.Context(), stationID, model.ParsePage(intParam(r, "limit", 25), intParam(r, "offset", 0)))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, values)
}

func (h *Handler) createTrap(w http.ResponseWriter, r *http.Request) {
	stationID, err := idParam(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var input model.Trap
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, err)
		return
	}
	input.StationID = stationID
	trap, err := h.services.Traps.Register(r.Context(), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, trap)
}

func (h *Handler) deployTrap(w http.ResponseWriter, r *http.Request) {
	h.changeTrapStatus(w, r, h.services.Traps.Deploy)
}

func (h *Handler) releaseTrap(w http.ResponseWriter, r *http.Request) {
	h.changeTrapStatus(w, r, h.services.Traps.Release)
}

func (h *Handler) breakTrap(w http.ResponseWriter, r *http.Request) {
	h.changeTrapStatus(w, r, h.services.Traps.Break)
}

func (h *Handler) changeTrapStatus(w http.ResponseWriter, r *http.Request, change func(context.Context, int64) error) {
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
