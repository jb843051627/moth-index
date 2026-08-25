package model

import (
	"errors"
	"math"
)

type Reading struct {
	ID          int64   `json:"id"`
	BatchID     int64   `json:"batch_id"`
	ObservedAt  string  `json:"observed_at"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Lux         float64 `json:"lux"`
	Rainfall    float64 `json:"rainfall"`
	Source      string  `json:"source"`
}

func (r Reading) Validate() error {
	if r.BatchID <= 0 {
		return errors.New("reading batch is required")
	}
	if r.ObservedAt == "" {
		return errors.New("reading time is required")
	}
	if math.IsNaN(r.Temperature) || math.IsInf(r.Temperature, 0) {
		return errors.New("temperature is not finite")
	}
	if r.Humidity < 0 || r.Humidity > 100 {
		return errors.New("humidity is outside range")
	}
	if r.Lux < 0 || r.Rainfall < 0 {
		return errors.New("reading values cannot be negative")
	}
	return nil
}

func (r Reading) IsWet() bool { return r.Rainfall >= 1.0 }

func (r Reading) IsWarm() bool { return r.Temperature >= 18 && r.Temperature <= 32 }

func (r Reading) QualityScore() float64 {
	score := 100.0
	if r.Humidity < 20 || r.Humidity > 95 {
		score -= 20
	}
	if r.Temperature < 5 || r.Temperature > 38 {
		score -= 25
	}
	if r.IsWet() {
		score -= 10
	}
	if score < 0 {
		return 0
	}
	return score
}

func (r Reading) Clone() Reading { return r }
