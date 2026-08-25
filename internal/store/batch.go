package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jb843051627/moth-index/internal/model"
)

type BatchStore struct{ db *DB }

func NewBatchStore(db *DB) *BatchStore { return &BatchStore{db: db} }

func (s *BatchStore) Create(ctx context.Context, batch model.NightBatch) (model.NightBatch, error) {
	result, err := s.db.Exec(ctx, `INSERT INTO batches(station_id,trap_id,started_at,ended_at,status,weather_note) VALUES(?,?,?,?,?,?)`,
		batch.StationID, batch.TrapID, batch.StartedAt, batch.EndedAt, batch.Status, batch.WeatherNote)
	if err != nil {
		return model.NightBatch{}, fmt.Errorf("insert batch: %w", err)
	}
	batch.ID, err = txInsertID(result)
	if err != nil {
		return model.NightBatch{}, err
	}
	return batch, nil
}

func (s *BatchStore) Get(ctx context.Context, id int64) (model.NightBatch, error) {
	var batch model.NightBatch
	err := s.db.QueryRow(ctx, `SELECT id,station_id,trap_id,started_at,ended_at,status,weather_note FROM batches WHERE id=?`, id).
		Scan(&batch.ID, &batch.StationID, &batch.TrapID, &batch.StartedAt, &batch.EndedAt, &batch.Status, &batch.WeatherNote)
	if err == sql.ErrNoRows {
		return model.NightBatch{}, ErrNotFound
	}
	if err != nil {
		return model.NightBatch{}, fmt.Errorf("get batch: %w", err)
	}
	return batch, nil
}

func (s *BatchStore) List(ctx context.Context, filter model.BatchFilter, limit, offset int) ([]model.NightBatch, error) {
	limit, offset = pageValues(limit, offset)
	query := `SELECT id,station_id,trap_id,started_at,ended_at,status,weather_note FROM batches WHERE 1=1`
	args := make([]any, 0, 6)
	if filter.StationID > 0 {
		query += ` AND station_id=?`
		args = append(args, filter.StationID)
	}
	if filter.Status != "" {
		query += ` AND status=?`
		args = append(args, filter.Status)
	}
	if filter.From != "" {
		query += ` AND started_at>=?`
		args = append(args, filter.From)
	}
	if filter.To != "" {
		query += ` AND started_at<=?`
		args = append(args, filter.To)
	}
	query += ` ORDER BY started_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list batches: %w", err)
	}
	values := make([]model.NightBatch, 0)
	for rows.Next() {
		var batch model.NightBatch
		if err := rows.Scan(&batch.ID, &batch.StationID, &batch.TrapID, &batch.StartedAt, &batch.EndedAt, &batch.Status, &batch.WeatherNote); err != nil {
			return nil, closeRows(rows, fmt.Errorf("scan batch: %w", err))
		}
		values = append(values, batch)
	}
	return values, closeRows(rows, nil)
}

func (s *BatchStore) SetStatus(ctx context.Context, id int64, status, endedAt string) error {
	result, err := s.db.Exec(ctx, `UPDATE batches SET status=?,ended_at=? WHERE id=?`, status, endedAt, id)
	if err != nil {
		return fmt.Errorf("set batch status: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read batch update: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *BatchStore) CountOpenByTrap(ctx context.Context, trapID int64) (int, error) {
	var count int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM batches WHERE trap_id=? AND status IN (?,?)`, trapID, model.BatchOpen, model.BatchReview).Scan(&count); err != nil {
		return 0, fmt.Errorf("count open batches: %w", err)
	}
	return count, nil
}
