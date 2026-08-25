package service

import (
	"context"
	"fmt"

	"github.com/jb843051627/moth-index/internal/model"
	"github.com/jb843051627/moth-index/internal/store"
)

type BatchService struct {
	batches  *store.BatchStore
	stations *store.StationStore
	traps    *store.TrapStore
	reports  *store.ReportStore
}

func NewBatchService(batches *store.BatchStore, stations *store.StationStore, traps *store.TrapStore, reports *store.ReportStore) *BatchService {
	return &BatchService{batches: batches, stations: stations, traps: traps, reports: reports}
}

func (s *BatchService) Open(ctx context.Context, input model.NightBatch) (model.NightBatch, error) {
	input.Status = model.BatchOpen
	input.EndedAt = ""
	if input.StartedAt == "" {
		input.StartedAt = serviceNow()
	}
	if err := input.Validate(); err != nil {
		return model.NightBatch{}, wrapValidation(err)
	}
	station, err := s.stations.Get(ctx, input.StationID)
	if err != nil {
		return model.NightBatch{}, err
	}
	if !station.IsUsable() {
		return model.NightBatch{}, fmt.Errorf("%w: station is retired", store.ErrState)
	}
	trap, err := s.traps.Get(ctx, input.TrapID)
	if err != nil {
		return model.NightBatch{}, err
	}
	if trap.StationID != input.StationID {
		return model.NightBatch{}, fmt.Errorf("%w: trap does not belong to station", store.ErrState)
	}
	if !trap.IsDeployable() {
		return model.NightBatch{}, fmt.Errorf("%w: trap cannot serve station", store.ErrState)
	}
	count, err := s.batches.CountOpenByTrap(ctx, input.TrapID)
	if err != nil {
		return model.NightBatch{}, err
	}
	if count > 0 {
		return model.NightBatch{}, fmt.Errorf("%w: trap already has open batch", store.ErrConflict)
	}
	return s.batches.Create(ctx, input)
}

func (s *BatchService) Get(ctx context.Context, id int64) (model.NightBatch, error) {
	return s.batches.Get(ctx, id)
}

func (s *BatchService) List(ctx context.Context, filter model.BatchFilter, page model.Page) ([]model.NightBatch, error) {
	if !filter.ValidWindow() {
		return nil, wrapValidation(fmt.Errorf("invalid time window"))
	}
	filter.Status = model.CleanStatus(filter.Status)
	page = page.Normalize()
	return s.batches.List(ctx, filter, page.Limit, page.Offset)
}

func (s *BatchService) StartReview(ctx context.Context, id int64) error {
	batch, err := s.batches.Get(ctx, id)
	if err != nil {
		return err
	}
	if !batch.CanReview() {
		return fmt.Errorf("%w: batch cannot enter review", store.ErrState)
	}
	return s.batches.SetStatus(ctx, id, model.BatchReview, "")
}

func (s *BatchService) Close(ctx context.Context, id int64) error {
	batch, err := s.batches.Get(ctx, id)
	if err != nil {
		return err
	}
	if !batch.CanClose() {
		return fmt.Errorf("%w: batch cannot close", store.ErrState)
	}
	ended := serviceNow()
	if err := s.batches.SetStatus(ctx, id, model.BatchClosed, ended); err != nil {
		return err
	}
	return s.reports.Audit(ctx, "batch", id, "close", ended)
}

func (s *BatchService) Archive(ctx context.Context, id int64) error {
	batch, err := s.batches.Get(ctx, id)
	if err != nil {
		return err
	}
	if !batch.CanArchive() {
		return fmt.Errorf("%w: batch is not closed", store.ErrState)
	}
	return s.batches.SetStatus(ctx, id, model.BatchArchived, batch.EndedAt)
}

func (s *BatchService) Duration(ctx context.Context, id int64) (int64, error) {
	batch, err := s.batches.Get(ctx, id)
	if err != nil {
		return 0, err
	}
	return int64(batch.Duration().Seconds()), nil
}
