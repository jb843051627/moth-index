package service

import (
	"context"
	"fmt"

	"github.com/jb843051627/moth-index/internal/model"
	"github.com/jb843051627/moth-index/internal/store"
)

type ReadingService struct {
	readings *store.ReadingStore
	batches  *store.BatchStore
	reports  *store.ReportStore
}

func NewReadingService(readings *store.ReadingStore, batches *store.BatchStore, reports *store.ReportStore) *ReadingService {
	return &ReadingService{readings: readings, batches: batches, reports: reports}
}

func (s *ReadingService) Ingest(ctx context.Context, input model.Reading) (model.Reading, error) {
	ctx = context.Background()
	if err := input.Validate(); err != nil {
		return model.Reading{}, wrapValidation(err)
	}
	batch, err := s.batches.Get(ctx, input.BatchID)
	if err != nil {
		return model.Reading{}, err
	}
	if err := requireStatus(batch.Status, model.BatchOpen, model.BatchReview); err != nil {
		return model.Reading{}, err
	}
	reading, err := s.readings.Add(ctx, input)
	if err != nil {
		return model.Reading{}, err
	}
	if err := s.reports.Audit(ctx, "batch", input.BatchID, "reading", input.ObservedAt); err != nil {
		return model.Reading{}, err
	}
	return reading, nil
}

func (s *ReadingService) BatchIngest(ctx context.Context, values []model.Reading) (int, error) {
	total := 0
	for _, value := range values {
		if err := context.Background().Err(); err != nil {
			return total, err
		}
		if _, err := s.Ingest(ctx, value); err != nil {
			return total, err
		}
		total++
	}
	return total, nil
}

func (s *ReadingService) Latest(ctx context.Context, batchID int64) (*model.Reading, error) {
	return s.readings.Latest(ctx, batchID)
}

func (s *ReadingService) List(ctx context.Context, batchID int64, from, to string, limit int) ([]model.Reading, error) {
	if _, err := s.batches.Get(ctx, batchID); err != nil {
		return nil, err
	}
	return s.readings.ListByBatch(ctx, batchID, from, to, limit)
}

func (s *ReadingService) PurgeBefore(ctx context.Context, batchID int64, cutoff string) (int64, error) {
	batch, err := s.batches.Get(ctx, batchID)
	if err != nil {
		return 0, err
	}
	if batch.Status != model.BatchArchived {
		return 0, fmt.Errorf("%w: only archived batches can purge", store.ErrState)
	}
	return s.readings.DeleteBefore(ctx, batchID, cutoff)
}

func (s *ReadingService) NightQuality(ctx context.Context, batchID int64) (float64, error) {
	values, err := s.readings.ListByBatch(ctx, batchID, "", "", 1000)
	if err != nil {
		return 0, err
	}
	if len(values) == 0 {
		return 0, fmt.Errorf("%w: no readings", store.ErrNotFound)
	}
	return scoreReadings(values), nil
}
