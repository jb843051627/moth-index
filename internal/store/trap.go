package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jb843051627/moth-index/internal/model"
)

type TrapStore struct{ db *DB }

func NewTrapStore(db *DB) *TrapStore { return &TrapStore{db: db} }

func (s *TrapStore) Create(ctx context.Context, trap model.Trap) (model.Trap, error) {
	result, err := s.db.Exec(ctx, `INSERT INTO traps(station_id,label,kind,status,installed_at) VALUES(?,?,?,?,?)`,
		trap.StationID, trap.Label, trap.Kind, trap.Status, trap.Installed)
	if err != nil {
		return model.Trap{}, fmt.Errorf("insert trap: %w", err)
	}
	trap.ID, err = txInsertID(result)
	if err != nil {
		return model.Trap{}, err
	}
	return trap, nil
}

func (s *TrapStore) Get(ctx context.Context, id int64) (model.Trap, error) {
	var trap model.Trap
	err := s.db.QueryRow(ctx, `SELECT id,station_id,label,kind,status,installed_at FROM traps WHERE id=?`, id).
		Scan(&trap.ID, &trap.StationID, &trap.Label, &trap.Kind, &trap.Status, &trap.Installed)
	if err == sql.ErrNoRows {
		return model.Trap{}, ErrNotFound
	}
	if err != nil {
		return model.Trap{}, fmt.Errorf("get trap: %w", err)
	}
	return trap, nil
}

func (s *TrapStore) ListByStation(ctx context.Context, stationID int64, limit, offset int) ([]model.Trap, error) {
	limit, offset = pageValues(limit, offset)
	rows, err := s.db.Query(ctx, `SELECT id,station_id,label,kind,status,installed_at FROM traps WHERE station_id=? ORDER BY label LIMIT ? OFFSET ?`, stationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list traps: %w", err)
	}
	values := make([]model.Trap, 0)
	for rows.Next() {
		var trap model.Trap
		if err := rows.Scan(&trap.ID, &trap.StationID, &trap.Label, &trap.Kind, &trap.Status, &trap.Installed); err != nil {
			return nil, closeRows(rows, fmt.Errorf("scan trap: %w", err))
		}
		values = append(values, trap)
	}
	return values, closeRows(rows, nil)
}

func (s *TrapStore) SetStatus(ctx context.Context, id int64, status string) error {
	result, err := s.db.Exec(ctx, `UPDATE traps SET status=? WHERE id=?`, status, id)
	if err != nil {
		return fmt.Errorf("set trap status: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read trap update: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *TrapStore) CountByStation(ctx context.Context, stationID int64) (int, error) {
	var count int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM traps WHERE station_id=?`, stationID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count traps: %w", err)
	}
	return count, nil
}

func (s *TrapStore) HasOpenBatch(ctx context.Context, trapID int64) (bool, error) {
	var count int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM batches WHERE trap_id=? AND status IN (?,?)`, trapID, model.BatchOpen, model.BatchReview).Scan(&count); err != nil {
		return false, fmt.Errorf("check trap batches: %w", err)
	}
	return count > 0, nil
}
