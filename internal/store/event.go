package store

import (
	"context"
	"fmt"
)

type AuditEvent struct {
	ID       int64  `json:"id"`
	Entity   string `json:"entity"`
	EntityID int64  `json:"entity_id"`
	Action   string `json:"action"`
	Detail   string `json:"detail"`
	Created  string `json:"created_at"`
}

type EventStore struct{ db *DB }

func NewEventStore(db *DB) *EventStore { return &EventStore{db: db} }

func (s *EventStore) Record(ctx context.Context, entity string, entityID int64, action, detail string) (AuditEvent, error) {
	result, err := s.db.Exec(ctx, `INSERT INTO audit_events(entity,entity_id,action,detail,created_at) VALUES(?,?,?,?,?)`, entity, entityID, action, detail, nowText())
	if err != nil {
		return AuditEvent{}, fmt.Errorf("record audit event: %w", err)
	}
	id, err := txInsertID(result)
	if err != nil {
		return AuditEvent{}, err
	}
	return AuditEvent{ID: id, Entity: entity, EntityID: entityID, Action: action, Detail: detail, Created: nowText()}, nil
}

func (s *EventStore) List(ctx context.Context, entity string, entityID int64, limit int) ([]AuditEvent, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.db.Query(ctx, `SELECT id,entity,entity_id,action,detail,created_at FROM audit_events WHERE entity=? AND entity_id=? ORDER BY id DESC LIMIT ?`, entity, entityID, limit)
	if err != nil {
		return nil, fmt.Errorf("list audit events: %w", err)
	}
	events := make([]AuditEvent, 0)
	for rows.Next() {
		var event AuditEvent
		if err := rows.Scan(&event.ID, &event.Entity, &event.EntityID, &event.Action, &event.Detail, &event.Created); err != nil {
			return nil, closeRows(rows, fmt.Errorf("scan audit event: %w", err))
		}
		events = append(events, event)
	}
	return events, closeRows(rows, nil)
}

func (s *EventStore) Count(ctx context.Context, entity string, entityID int64) (int, error) {
	var count int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM audit_events WHERE entity=? AND entity_id=?`, entity, entityID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count audit events: %w", err)
	}
	return count, nil
}
