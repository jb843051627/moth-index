package model

import "errors"

const (
	SeverityInfo    = "info"
	SeverityWarn    = "warn"
	SeverityBlocker = "blocker"
)

type Signal struct {
	ID        int64  `json:"id"`
	BatchID   int64  `json:"batch_id"`
	Code      string `json:"code"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"created_at"`
}

func (s Signal) Validate() error {
	if s.BatchID <= 0 || s.Code == "" {
		return errors.New("signal batch and code are required")
	}
	if s.Severity != SeverityInfo && s.Severity != SeverityWarn && s.Severity != SeverityBlocker {
		return errors.New("unknown signal severity")
	}
	return nil
}

func (s Signal) BlocksReview() bool { return s.Active && s.Severity == SeverityWarn }

func (s Signal) Clone() Signal { return s }

func SeverityRank(value string) int {
	switch value {
	case SeverityBlocker:
		return 3
	case SeverityWarn:
		return 2
	case SeverityInfo:
		return 1
	default:
		return 0
	}
}
