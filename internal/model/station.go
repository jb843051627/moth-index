package model

import "errors"

const (
	StationActive  = "active"
	StationRetired = "retired"
)

type Station struct {
	ID        int64  `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Habitat   string `json:"habitat"`
	Timezone  string `json:"timezone"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func (s Station) Validate() error {
	if s.Code == "" {
		return errors.New("station code is required")
	}
	if s.Name == "" {
		return errors.New("station name is required")
	}
	if s.Timezone == "" {
		return errors.New("station timezone is required")
	}
	return nil
}

func (s Station) CanRetire() bool { return s.Status == StationActive }

func (s Station) IsUsable() bool { return s.Status == StationActive && s.Code != "" }

func (s Station) Clone() Station { return s }

func NormalizeStationCode(value string) string {
	out := make([]rune, 0, len(value))
	for _, r := range value {
		if r >= 'a' && r <= 'z' {
			out = append(out, r-'a'+'A')
			continue
		}
		if r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' {
			out = append(out, r)
		}
	}
	return string(out)
}

func StationStatusLabel(status string) string {
	switch status {
	case StationActive:
		return "active"
	case StationRetired:
		return "retired"
	default:
		return "unknown"
	}
}
