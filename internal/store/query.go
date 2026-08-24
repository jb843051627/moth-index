package store

import (
	"context"
	"fmt"
)

type QueryStore struct{ db *DB }

func NewQueryStore(db *DB) *QueryStore { return &QueryStore{db: db} }

func (s *QueryStore) CountByStatus(ctx context.Context, table, status string) (int, error) {
	allowed := map[string]bool{"stations": true, "traps": true, "batches": true, "specimens": true, "review_tasks": true}
	if !allowed[table] {
		return 0, fmt.Errorf("unsupported status table")
	}
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE status=?", table)
	var count int
	if err := s.db.QueryRow(ctx, query, status).Scan(&count); err != nil {
		return 0, fmt.Errorf("count status: %w", err)
	}
	return count, nil
}

func (s *QueryStore) LatestEvent(ctx context.Context, entity string, id int64) (AuditEvent, error) {
	var event AuditEvent
	err := s.db.QueryRow(ctx, `SELECT id,entity,entity_id,action,detail,created_at FROM audit_events WHERE entity=? AND entity_id=? ORDER BY id DESC LIMIT 1`, entity, id).
		Scan(&event.ID, &event.Entity, &event.EntityID, &event.Action, &event.Detail, &event.Created)
	if err != nil {
		if isNoRows(err) {
			return AuditEvent{}, ErrNotFound
		}
		return AuditEvent{}, fmt.Errorf("latest event: %w", err)
	}
	return event, nil
}

func (s *QueryStore) BatchIDsWithActiveBlockers(ctx context.Context, limit int) ([]int64, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.Query(ctx, `SELECT DISTINCT batch_id FROM signals WHERE active=1 AND severity=? ORDER BY batch_id LIMIT ?`, "blocker", limit)
	if err != nil {
		return nil, fmt.Errorf("list blocked batches: %w", err)
	}
	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, closeRows(rows, fmt.Errorf("scan blocked batch: %w", err))
		}
		ids = append(ids, id)
	}
	return ids, closeRows(rows, nil)
}

func (s *QueryStore) ReadingCount(ctx context.Context, batchID int64) (int, error) {
	var count int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM readings WHERE batch_id=?`, batchID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count readings: %w", err)
	}
	return count, nil
}

func (s *QueryStore) SpecimenTags(ctx context.Context, batchID int64) ([]string, error) {
	rows, err := s.db.Query(ctx, `SELECT tag FROM specimens WHERE batch_id=? ORDER BY tag`, batchID)
	if err != nil {
		return nil, fmt.Errorf("list specimen tags: %w", err)
	}
	tags := make([]string, 0)
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, closeRows(rows, fmt.Errorf("scan specimen tag: %w", err))
		}
		tags = append(tags, tag)
	}
	return tags, closeRows(rows, nil)
}

func (s *QueryStore) DeleteBatchSignals(ctx context.Context, batchID int64) (int64, error) {
	result, err := s.db.Exec(ctx, `DELETE FROM signals WHERE batch_id=?`, batchID)
	if err != nil {
		return 0, fmt.Errorf("delete batch signals: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("read signal deletion: %w", err)
	}
	return count, nil
}
