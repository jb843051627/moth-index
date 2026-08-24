package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jb843051627/moth-index/internal/model"
)

type ReadingStore struct{ db *DB }

func NewReadingStore(db *DB) *ReadingStore { return &ReadingStore{db: db} }

func (s *ReadingStore) Add(ctx context.Context, reading model.Reading) (model.Reading, error) {
	result, err := s.db.Exec(ctx, `INSERT INTO readings(batch_id,observed_at,temperature,humidity,lux,rainfall,source) VALUES(?,?,?,?,?,?,?)`,
		reading.BatchID, reading.ObservedAt, reading.Temperature, reading.Humidity, reading.Lux, reading.Rainfall, reading.Source)
	if err != nil {
		return model.Reading{}, fmt.Errorf("insert reading: %w", err)
	}
	reading.ID, err = txInsertID(result)
	if err != nil {
		return model.Reading{}, err
	}
	return reading, nil
}

func (s *ReadingStore) Latest(ctx context.Context, batchID int64) (*model.Reading, error) {
	var reading model.Reading
	err := s.db.QueryRow(ctx, `SELECT id,batch_id,observed_at,temperature,humidity,lux,rainfall,source FROM readings WHERE batch_id=? ORDER BY observed_at DESC,id DESC LIMIT 1`, batchID).
		Scan(&reading.ID, &reading.BatchID, &reading.ObservedAt, &reading.Temperature, &reading.Humidity, &reading.Lux, &reading.Rainfall, &reading.Source)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get latest reading: %w", err)
	}
	return &reading, nil
}

func (s *ReadingStore) ListByBatch(ctx context.Context, batchID int64, from, to string, limit int) ([]model.Reading, error) {
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	query := `SELECT id,batch_id,observed_at,temperature,humidity,lux,rainfall,source FROM readings WHERE batch_id=?`
	args := []any{batchID}
	if from != "" {
		query += ` AND observed_at>=?`
		args = append(args, from)
	}
	if to != "" {
		query += ` AND observed_at<=?`
		args = append(args, to)
	}
	query += ` ORDER BY observed_at DESC,id DESC LIMIT ?`
	args = append(args, limit)
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list readings: %w", err)
	}
	values := make([]model.Reading, 0)
	for rows.Next() {
		var reading model.Reading
		if err := rows.Scan(&reading.ID, &reading.BatchID, &reading.ObservedAt, &reading.Temperature, &reading.Humidity, &reading.Lux, &reading.Rainfall, &reading.Source); err != nil {
			return nil, closeRows(rows, fmt.Errorf("scan reading: %w", err))
		}
		values = append(values, reading)
	}
	return values, closeRows(rows, nil)
}

func (s *ReadingStore) DeleteBefore(ctx context.Context, batchID int64, cutoff string) (int64, error) {
	result, err := s.db.Exec(ctx, `DELETE FROM readings WHERE batch_id=? AND observed_at<?`, batchID, cutoff)
	if err != nil {
		return 0, fmt.Errorf("delete readings: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("read deleted readings: %w", err)
	}
	return count, nil
}
