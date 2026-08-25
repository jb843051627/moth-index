package store

import (
	"context"
	"fmt"
)

type MaintenanceStore struct{ db *DB }

func NewMaintenanceStore(db *DB) *MaintenanceStore { return &MaintenanceStore{db: db} }

func (s *MaintenanceStore) PruneEvents(ctx context.Context, cutoff string) (int64, error) {
	result, err := s.db.Exec(ctx, `DELETE FROM audit_events WHERE created_at<?`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("prune audit events: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("read pruned events: %w", err)
	}
	return count, nil
}

func (s *MaintenanceStore) ResetRunningTasks(ctx context.Context) (int64, error) {
	result, err := s.db.Exec(ctx, `UPDATE review_tasks SET status=?,updated_at=?,error_text=? WHERE status=?`, "queued", nowText(), "worker restarted", "running")
	if err != nil {
		return 0, fmt.Errorf("reset running tasks: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("read reset tasks: %w", err)
	}
	return count, nil
}

func (s *MaintenanceStore) DeleteFailedTasks(ctx context.Context, before string) (int64, error) {
	result, err := s.db.Exec(ctx, `DELETE FROM review_tasks WHERE status=? AND updated_at<?`, "failed", before)
	if err != nil {
		return 0, fmt.Errorf("delete failed tasks: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("read deleted tasks: %w", err)
	}
	return count, nil
}

func (s *MaintenanceStore) Vacuum(ctx context.Context) error {
	if _, err := s.db.Exec(ctx, `PRAGMA optimize`); err != nil {
		return fmt.Errorf("optimize sqlite: %w", err)
	}
	return nil
}
