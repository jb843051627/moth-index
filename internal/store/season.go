package store

import (
	"context"
	"fmt"
)

type ReadingStats struct {
	BatchID int64
	Count   int
	Mean    float64
	Low     float64
	High    float64
}

type SeasonStore struct{ db *DB }

func NewSeasonStore(db *DB) *SeasonStore { return &SeasonStore{db: db} }

func (s *SeasonStore) Stats(ctx context.Context, batchID int64) (ReadingStats, error) {
	var stats ReadingStats
	stats.BatchID = batchID
	err := s.db.QueryRow(ctx, `SELECT COUNT(*),COALESCE(AVG(temperature),0),COALESCE(MIN(temperature),0),COALESCE(MAX(temperature),0) FROM readings WHERE batch_id=?`, batchID).
		Scan(&stats.Count, &stats.Mean, &stats.Low, &stats.High)
	if err != nil {
		return ReadingStats{}, fmt.Errorf("reading stats: %w", err)
	}
	return stats, nil
}

func (s *SeasonStore) BatchStatus(ctx context.Context, ids []int64) (map[int64]string, error) {
	result := make(map[int64]string, len(ids))
	for _, id := range ids {
		var status string
		if err := s.db.QueryRow(ctx, `SELECT status FROM batches WHERE id=?`, id).Scan(&status); err != nil {
			return nil, fmt.Errorf("batch status %d: %w", id, err)
		}
		result[id] = status
	}
	return result, nil
}

func (s *SeasonStore) MarkSignalsResolved(ctx context.Context, batchID int64) (int64, error) {
	result, err := s.db.Exec(ctx, `UPDATE signals SET active=0 WHERE batch_id=? AND active=1`, batchID)
	if err != nil {
		return 0, fmt.Errorf("resolve batch signals: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("read resolved signals: %w", err)
	}
	return count, nil
}

func (s *SeasonStore) DistinctStations(ctx context.Context, status string, limit int) ([]int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.db.Query(ctx, `SELECT DISTINCT station_id FROM batches WHERE status=? ORDER BY station_id LIMIT ?`, status, limit)
	if err != nil {
		return nil, fmt.Errorf("distinct stations: %w", err)
	}
	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, closeRows(rows, fmt.Errorf("scan station id: %w", err))
		}
		ids = append(ids, id)
	}
	return ids, closeRows(rows, nil)
}
