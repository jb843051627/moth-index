package model

import "errors"

type Taxon struct {
	ID        int64  `json:"id"`
	Family    string `json:"family"`
	Genus     string `json:"genus"`
	Species   string `json:"species"`
	Common    string `json:"common_name"`
	Authority string `json:"authority"`
}

type Classification struct {
	SpecimenID int64   `json:"specimen_id"`
	TaxonID    int64   `json:"taxon_id"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"`
}

func (t Taxon) Validate() error {
	if t.Family == "" || t.Genus == "" || t.Species == "" {
		return errors.New("taxon family, genus and species are required")
	}
	return nil
}

func (t Taxon) ScientificName() string { return t.Genus + " " + t.Species }

func (c Classification) Validate() error {
	if c.SpecimenID <= 0 || c.TaxonID <= 0 {
		return errors.New("classification references are required")
	}
	if c.Confidence < 0 || c.Confidence > 1 {
		return errors.New("classification confidence is outside range")
	}
	return nil
}

func ConfidenceBand(value float64) string {
	if value >= 0.9 {
		return "high"
	}
	if value >= 0.65 {
		return "medium"
	}
	return "low"
}

func (t Taxon) Clone() Taxon { return t }

func (c Classification) Clone() Classification { return c }
