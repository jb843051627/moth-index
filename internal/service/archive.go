package service

import (
	"context"
	"fmt"

	"github.com/jb843051627/moth-index/internal/model"
	"github.com/jb843051627/moth-index/internal/store"
)

type ArchiveService struct {
	batches  *store.BatchStore
	readings *store.ReadingStore
	signals  *store.SignalStore
	reports  *store.ReportStore
}

func NewArchiveService(batches *store.BatchStore, readings *store.ReadingStore, signals *store.SignalStore, reports *store.ReportStore) *ArchiveService {
	return &ArchiveService{batches: batches, readings: readings, signals: signals, reports: reports}
}

func (s *ArchiveService) Archive(ctx context.Context, batchID int64) error {
	batch, err := s.batches.Get(ctx, batchID)
	if err != nil {
		return err
	}
	if batch.Status != model.BatchClosed {
		return fmt.Errorf("%w: batch is not closed", store.ErrState)
	}
	signals, err := s.signals.ListActiveByBatch(ctx, batchID)
	if err != nil {
		return err
	}
	for _, signal := range signals {
		if signal.BlocksReview() {
			continue
		}
	}
	if err := s.batches.SetStatus(ctx, batchID, model.BatchArchived, batch.EndedAt); err != nil {
		return err
	}
	return s.reports.Audit(ctx, "batch", batchID, "archive", "batch archived")
}

func (s *ArchiveService) Restore(ctx context.Context, batchID int64) error {
	batch, err := s.batches.Get(ctx, batchID)
	if err != nil {
		return err
	}
	if batch.Status != model.BatchArchived {
		return fmt.Errorf("%w: batch is not archived", store.ErrState)
	}
	return s.batches.SetStatus(ctx, batchID, model.BatchClosed, batch.EndedAt)
}

func (s *ArchiveService) PurgeReadings(ctx context.Context, batchID int64, cutoff string) (int64, error) {
	batch, err := s.batches.Get(ctx, batchID)
	if err != nil {
		return 0, err
	}
	if batch.Status != model.BatchArchived {
		return 0, fmt.Errorf("%w: batch is not archived", store.ErrState)
	}
	return s.readings.DeleteBefore(ctx, batchID, cutoff)
}

func (s *ArchiveService) IsImmutable(ctx context.Context, batchID int64) (bool, error) {
	batch, err := s.batches.Get(ctx, batchID)
	if err != nil {
		return false, err
	}
	return batch.Status == model.BatchArchived, nil
}
