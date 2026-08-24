package handler

import (
	"net/http"

	"github.com/jb843051627/moth-index/internal/model"
)

func (h *Handler) listSpecimens(w http.ResponseWriter, r *http.Request) {
	batchID, err := idParam(r)
	if err != nil {
		writeError(w, err)
		return
	}
	values, err := h.services.Specimens.List(r.Context(), batchID, model.ParsePage(intParam(r, "limit", 25), intParam(r, "offset", 0)))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, values)
}

func (h *Handler) captureSpecimen(w http.ResponseWriter, r *http.Request) {
	batchID, err := idParam(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var input model.Specimen
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, err)
		return
	}
	input.BatchID = batchID
	specimen, err := h.services.Specimens.Capture(r.Context(), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, specimen)
}

func (h *Handler) getSpecimen(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		writeError(w, err)
		return
	}
	specimen, err := h.services.Specimens.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, specimen)
}

func (h *Handler) classifySpecimen(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var taxon model.Taxon
	if err := decodeJSON(r, &taxon); err != nil {
		writeError(w, err)
		return
	}
	specimen, err := h.services.Specimens.Classify(r.Context(), id, taxon)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, specimen)
}

func (h *Handler) approveSpecimen(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var input struct {
		Reviewer   string  `json:"reviewer"`
		Confidence float64 `json:"confidence"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, err)
		return
	}
	review, err := h.services.Reviews.Approve(r.Context(), id, input.Reviewer, input.Confidence)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, review)
}

func (h *Handler) rejectSpecimen(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var input struct {
		Reviewer string `json:"reviewer"`
		Note     string `json:"note"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, err)
		return
	}
	review, err := h.services.Reviews.Reject(r.Context(), id, input.Reviewer, input.Note)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, review)
}
