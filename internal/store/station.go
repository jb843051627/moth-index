package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jb843051627/moth-index/internal/model"
)

type StationStore struct{ db *DB }

func NewStationStore(db *DB) *StationStore { return &StationStore{db: db} }

func (s *StationStore) Create(ctx context.Context, station model.Station) (model.Station, error) {
	result, err := s.db.Exec(ctx, `INSERT INTO stations(code,name,habitat,timezone,status,created_at) VALUES(?,?,?,?,?,?)`,
		station.Code, station.Name, station.Habitat, station.Timezone, station.Status, station.CreatedAt)
	if err != nil {
		return model.Station{}, fmt.Errorf("insert station: %w", err)
	}
	station.ID, err = txInsertID(result)
	if err != nil {
		return model.Station{}, err
	}
	return station, nil
}

func (s *StationStore) Get(ctx context.Context, id int64) (model.Station, error) {
	var station model.Station
	err := s.db.QueryRow(ctx, `SELECT id,code,name,habitat,timezone,status,created_at FROM stations WHERE id=?`, id).
		Scan(&station.ID, &station.Code, &station.Name, &station.Habitat, &station.Timezone, &station.Status, &station.CreatedAt)
	if err == sql.ErrNoRows {
		return model.Station{}, ErrNotFound
	}
	if err != nil {
		return model.Station{}, fmt.Errorf("get station: %w", err)
	}
	return station, nil
}

func (s *StationStore) GetByCode(ctx context.Context, code string) (model.Station, error) {
	var station model.Station
	err := s.db.QueryRow(ctx, `SELECT id,code,name,habitat,timezone,status,created_at FROM stations WHERE code=?`, code).
		Scan(&station.ID, &station.Code, &station.Name, &station.Habitat, &station.Timezone, &station.Status, &station.CreatedAt)
	if err == sql.ErrNoRows {
		return model.Station{}, ErrNotFound
	}
	if err != nil {
		return model.Station{}, fmt.Errorf("get station by code: %w", err)
	}
	return station, nil
}

func (s *StationStore) List(ctx context.Context, limit, offset int, includeRetired bool) ([]model.Station, error) {
	limit, offset = pageValues(limit, offset)
	query := `SELECT id,code,name,habitat,timezone,status,created_at FROM stations ORDER BY code LIMIT ? OFFSET ?`
	args := []any{limit, offset}
	if !includeRetired {
		query = `SELECT id,code,name,habitat,timezone,status,created_at FROM stations WHERE status=? ORDER BY code LIMIT ? OFFSET ?`
		args = []any{model.StationActive, limit, offset}
	}
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list stations: %w", err)
	}
	values := make([]model.Station, 0)
	for rows.Next() {
		var station model.Station
		if err := rows.Scan(&station.ID, &station.Code, &station.Name, &station.Habitat, &station.Timezone, &station.Status, &station.CreatedAt); err != nil {
			return nil, closeRows(rows, fmt.Errorf("scan station: %w", err))
		}
		values = append(values, station)
	}
	return values, closeRows(rows, nil)
}

func (s *StationStore) UpdateStatus(ctx context.Context, id int64, status string) error {
	result, err := s.db.Exec(ctx, `UPDATE stations SET status=? WHERE id=?`, status, id)
	if err != nil {
		return fmt.Errorf("update station status: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read station update: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *StationStore) Count(ctx context.Context) (int, error) {
	var count int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM stations`).Scan(&count); err != nil {
		return 0, fmt.Errorf("count stations: %w", err)
	}
	return count, nil
}
