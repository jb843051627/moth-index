package model

import (
	"errors"
	"strings"
)

const (
	SpecimenCaptured = "captured"
	SpecimenPending  = "pending"
	SpecimenAccepted = "accepted"
	SpecimenRejected = "rejected"
)

type Specimen struct {
	ID        int64  `json:"id"`
	BatchID   int64  `json:"batch_id"`
	Tag       string `json:"tag"`
	Family    string `json:"family"`
	Genus     string `json:"genus"`
	Species   string `json:"species"`
	Sex       string `json:"sex"`
	Count     int    `json:"count"`
	Status    string `json:"status"`
	Notes     string `json:"notes"`
	CreatedAt string `json:"created_at"`
}

func (s Specimen) Validate() error {
	if s.BatchID <= 0 {
		return errors.New("specimen batch is required")
	}
	if NormalizeTag(s.Tag) == "" {
		return errors.New("specimen tag is required")
	}
	if s.Count <= 0 {
		return errors.New("specimen count must be positive")
	}
	return nil
}

func NormalizeTag(tag string) string { return strings.ToUpper(strings.TrimSpace(tag)) }

func (s Specimen) ScientificName() string {
	parts := []string{s.Genus, s.Species}
	value := strings.TrimSpace(strings.Join(parts, " "))
	if value == "" {
		return "unidentified"
	}
	return value
}

func (s Specimen) IsClassified() bool { return s.Family != "" && s.Genus != "" && s.Species != "" }

func (s Specimen) CanReview() bool {
	return s.Status == SpecimenCaptured || s.Status == SpecimenPending
}

func (s Specimen) Clone() Specimen { return s }

func SpecimenStatusLabel(status string) string {
	switch status {
	case SpecimenCaptured:
		return "captured"
	case SpecimenPending:
		return "pending"
	case SpecimenAccepted:
		return "accepted"
	case SpecimenRejected:
		return "rejected"
	default:
		return "unknown"
	}
}
