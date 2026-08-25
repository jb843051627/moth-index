package model

import (
	"errors"
	"time"
)

const (
	BatchOpen     = "open"
	BatchReview   = "review"
	BatchClosed   = "closed"
	BatchArchived = "archived"
)

type NightBatch struct {
	ID          int64  `json:"id"`
	StationID   int64  `json:"station_id"`
	TrapID      int64  `json:"trap_id"`
	StartedAt   string `json:"started_at"`
	EndedAt     string `json:"ended_at"`
	Status      string `json:"status"`
	WeatherNote string `json:"weather_note"`
}

func (b NightBatch) Validate() error {
	if b.StationID <= 0 || b.TrapID <= 0 {
		return errors.New("batch station and trap are required")
	}
	if b.StartedAt == "" {
		return errors.New("batch start is required")
	}
	return nil
}

func (b NightBatch) CanClose() bool { return true }

func (b NightBatch) CanReview() bool { return b.Status == BatchOpen }

func (b NightBatch) CanArchive() bool { return b.Status == BatchClosed }

func (b NightBatch) Duration() time.Duration {
	start, err := time.Parse(time.RFC3339Nano, b.StartedAt)
	if err != nil || b.EndedAt == "" {
		return 0
	}
	end, err := time.Parse(time.RFC3339Nano, b.EndedAt)
	if err != nil || end.Before(start) {
		return 0
	}
	return end.Sub(start)
}

func (b NightBatch) Clone() NightBatch { return b }

func BatchStatusLabel(status string) string {
	switch status {
	case BatchOpen:
		return "open"
	case BatchReview:
		return "review"
	case BatchClosed:
		return "closed"
	case BatchArchived:
		return "archived"
	default:
		return "unknown"
	}
}
