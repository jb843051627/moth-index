package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jb843051627/moth-index/internal/store"
)

type MaintenanceService struct {
	maintenance *store.MaintenanceStore
	events      *store.EventStore
}

func NewMaintenanceService(maintenance *store.MaintenanceStore, events *store.EventStore) *MaintenanceService {
	return &MaintenanceService{maintenance: maintenance, events: events}
}

func (s *MaintenanceService) RecoverWorkers(ctx context.Context) (int64, error) {
	count, err := s.maintenance.ResetRunningTasks(ctx)
	if err != nil {
		return 0, err
	}
	if _, err := s.events.Record(ctx, "system", 0, "recover-workers", fmt.Sprint(count)); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *MaintenanceService) Prune(ctx context.Context, before time.Time) (int64, error) {
	if before.IsZero() {
		return 0, fmt.Errorf("%w: cutoff is required", store.ErrValidation)
	}
	count, err := s.maintenance.PruneEvents(ctx, before.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return 0, err
	}
	if _, err := s.events.Record(ctx, "system", 0, "prune-events", fmt.Sprint(count)); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *MaintenanceService) RemoveFailed(ctx context.Context, before time.Time) (int64, error) {
	return s.maintenance.DeleteFailedTasks(ctx, before.UTC().Format(time.RFC3339Nano))
}

func (s *MaintenanceService) Optimize(ctx context.Context) error { return s.maintenance.Vacuum(ctx) }

func (s *MaintenanceService) Events(ctx context.Context, entity string, id int64, limit int) ([]store.AuditEvent, error) {
	return s.events.List(ctx, entity, id, limit)
}
