package model

import "errors"

const (
	ReviewPending  = "pending"
	ReviewApproved = "approved"
	ReviewRejected = "rejected"
)

type Review struct {
	ID         int64   `json:"id"`
	SpecimenID int64   `json:"specimen_id"`
	Reviewer   string  `json:"reviewer"`
	Decision   string  `json:"decision"`
	Confidence float64 `json:"confidence"`
	Note       string  `json:"note"`
	ReviewedAt string  `json:"reviewed_at"`
}

func (r Review) Validate() error {
	if r.SpecimenID <= 0 || r.Reviewer == "" {
		return errors.New("review specimen and reviewer are required")
	}
	if r.Decision != ReviewPending && r.Decision != ReviewApproved && r.Decision != ReviewRejected {
		return errors.New("unknown review decision")
	}
	if r.Confidence < 0 || r.Confidence > 1 {
		return errors.New("review confidence is outside range")
	}
	return nil
}

func (r Review) IsFinal() bool { return r.Decision == ReviewApproved || r.Decision == ReviewRejected }

func (r Review) Clone() Review { return r }

func ReviewDecisionLabel(value string) string {
	switch value {
	case ReviewApproved:
		return "approved"
	case ReviewRejected:
		return "rejected"
	default:
		return "pending"
	}
}
