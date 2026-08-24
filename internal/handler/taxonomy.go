package handler

import (
	"net/http"

	"github.com/jb843051627/moth-index/internal/model"
)

func (h *Handler) searchTaxa(w http.ResponseWriter, r *http.Request) {
	values, err := h.services.Taxonomy.Search(r.Context(), r.URL.Query().Get("q"), intParam(r, "limit", 20))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, values)
}

func (h *Handler) registerTaxon(w http.ResponseWriter, r *http.Request) {
	var input model.Taxon
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, err)
		return
	}
	taxon, err := h.services.Taxonomy.Register(r.Context(), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, taxon)
}
