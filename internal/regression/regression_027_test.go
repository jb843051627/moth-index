package regression

import (
	"testing"

	"github.com/jb843051627/moth-index/internal/model"
)

func TestBug27_MidnightFlightWindowUsesShortArc(t *testing.T) {
	values := []model.Reading{{ObservedAt: "2026-08-24T23:00:00Z"}, {ObservedAt: "2026-08-25T01:00:00Z"}}
	window := model.EstimateFlightWindow(values)
	if window.Duration != 2 { t.Fatalf("duration=%v, want 2", window.Duration) }
}
