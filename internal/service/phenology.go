package service

import (
	"context"
	"fmt"

	"github.com/jb843051627/moth-index/internal/model"
	"github.com/jb843051627/moth-index/internal/store"
)

type PhenologyReport struct {
	BatchID        int64                `json:"batch_id"`
	Flight         model.FlightWindow   `json:"flight"`
	LowTemp        float64              `json:"low_temp"`
	HighTemp       float64              `json:"high_temp"`
	HumiditySwing  float64              `json:"humidity_swing"`
	CaptureRate    float64              `json:"capture_rate"`
	ExpectedEffort int                  `json:"expected_effort"`
	RainBiased     bool                 `json:"rain_biased"`
	Profile        model.CaptureProfile `json:"profile"`
}

type PhenologyService struct {
	batches   *store.BatchStore
	readings  *store.ReadingStore
	specimens *store.SpecimenStore
	traps     *store.TrapStore
}

func NewPhenologyService(batches *store.BatchStore, readings *store.ReadingStore, specimens *store.SpecimenStore, traps *store.TrapStore) *PhenologyService {
	return &PhenologyService{batches: batches, readings: readings, specimens: specimens, traps: traps}
}

func (s *PhenologyService) Report(ctx context.Context, batchID int64) (PhenologyReport, error) {
	batch, err := s.batches.Get(ctx, batchID)
	if err != nil {
		return PhenologyReport{}, err
	}
	readings, err := s.readings.ListByBatch(ctx, batchID, "", "", 1000)
	if err != nil {
		return PhenologyReport{}, err
	}
	specimens, err := s.specimens.ListByBatch(ctx, batchID, 1000, 0)
	if err != nil {
		return PhenologyReport{}, err
	}
	traps, err := s.traps.ListByStation(ctx, batch.StationID, 100, 0)
	if err != nil {
		return PhenologyReport{}, err
	}
	flight := model.EstimateFlightWindow(readings)
	low, high := model.NightTemperatureRange(readings)
	report := PhenologyReport{BatchID: batchID, Flight: flight, LowTemp: low, HighTemp: high, HumiditySwing: model.HumiditySwing(readings), CaptureRate: model.CaptureRate(specimens, flight.Duration), ExpectedEffort: model.ExpectedEffort(flight, len(traps)), RainBiased: model.IsLikelyRainBiased(readings), Profile: model.BuildCaptureProfile(specimens)}
	if len(readings) > 0 && len(specimens) == 0 {
		return report, fmt.Errorf("%w: samples have no environment context", store.ErrState)
	}
	return report, nil
}

func (s *PhenologyService) Compare(ctx context.Context, firstID, secondID int64) (map[string]float64, error) {
	first, err := s.Report(ctx, firstID)
	if err != nil {
		return nil, err
	}
	second, err := s.Report(ctx, secondID)
	if err != nil {
		return nil, err
	}
	return map[string]float64{
		"capture_rate_delta": second.CaptureRate - first.CaptureRate,
		"temperature_delta":  (second.HighTemp + second.LowTemp - first.HighTemp - first.LowTemp) / 2,
		"humidity_delta":     second.HumiditySwing - first.HumiditySwing,
		"diversity_delta":    second.Profile.Diversity - first.Profile.Diversity,
	}, nil
}

func (s *PhenologyService) IsComparable(ctx context.Context, firstID, secondID int64) (bool, error) {
	first, err := s.batches.Get(ctx, firstID)
	if err != nil {
		return false, err
	}
	second, err := s.batches.Get(ctx, secondID)
	if err != nil {
		return false, err
	}
	return first.StationID == second.StationID && first.TrapID == second.TrapID, nil
}

func (s *PhenologyService) CaptureRate(ctx context.Context, batchID int64) (float64, error) {
	report, err := s.Report(ctx, batchID)
	if err != nil {
		return 0, err
	}
	return report.CaptureRate, nil
}
