package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jb843051627/moth-index/internal/model"
	"github.com/jb843051627/moth-index/internal/store"
)

type SeasonalService struct {
	batches   *store.BatchStore
	readings  *store.ReadingStore
	specimens *store.SpecimenStore
	traps     *store.TrapStore
	seasons   *store.SeasonStore
}

func NewSeasonalService(batches *store.BatchStore, readings *store.ReadingStore, specimens *store.SpecimenStore, traps *store.TrapStore, seasons *store.SeasonStore) *SeasonalService {
	return &SeasonalService{batches: batches, readings: readings, specimens: specimens, traps: traps, seasons: seasons}
}

func (s *SeasonalService) Score(ctx context.Context, batchID int64) (float64, error) {
	batch, err := s.batches.Get(ctx, batchID)
	if err != nil {
		return 0, err
	}
	readings, err := s.readings.ListByBatch(ctx, batchID, "", "", 1000)
	if err != nil {
		return 0, err
	}
	start, err := time.Parse(time.RFC3339Nano, batch.StartedAt)
	if err != nil {
		return 0, fmt.Errorf("parse season start: %w", err)
	}
	return model.EmergenceScore(readings, start), nil
}

func (s *SeasonalService) Stats(ctx context.Context, batchID int64) (store.ReadingStats, error) {
	return s.seasons.Stats(ctx, batchID)
}

func (s *SeasonalService) Effort(ctx context.Context, batchID int64) (int, error) {
	batch, err := s.batches.Get(ctx, batchID)
	if err != nil {
		return 0, err
	}
	readings, err := s.readings.ListByBatch(ctx, batchID, "", "", 1000)
	if err != nil {
		return 0, err
	}
	traps, err := s.traps.ListByStation(ctx, batch.StationID, 100, 0)
	if err != nil {
		return 0, err
	}
	return model.ExpectedEffort(model.EstimateFlightWindow(readings), len(traps)), nil
}

func (s *SeasonalService) CaptureRate(ctx context.Context, batchID int64) (float64, error) {
	readings, err := s.readings.ListByBatch(ctx, batchID, "", "", 1000)
	if err != nil {
		return 0, err
	}
	specimens, err := s.specimens.ListByBatch(ctx, batchID, 1000, 0)
	if err != nil {
		return 0, err
	}
	window := model.EstimateFlightWindow(readings)
	return model.CaptureRate(specimens, window.Duration), nil
}

func (s *SeasonalService) ValidateSurveyNight(ctx context.Context, batchID int64) error {
	batch, err := s.batches.Get(ctx, batchID)
	if err != nil {
		return err
	}
	start, err := time.Parse(time.RFC3339Nano, batch.StartedAt)
	if err != nil {
		return err
	}
	if !model.IsSurveyNight(start) {
		return fmt.Errorf("%w: batch started outside night survey window", store.ErrState)
	}
	return nil
}
