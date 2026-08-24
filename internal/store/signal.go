package store

import (
	"context"
	"fmt"

	"github.com/jb843051627/moth-index/internal/model"
)

type SignalStore struct{ db *DB }

func NewSignalStore(db *DB) *SignalStore { return &SignalStore{db: db} }

func (s *SignalStore) Create(ctx context.Context, signal model.Signal) (model.Signal, error) {
	active := 0
	if signal.Active {
		active = 1
	}
	result, err := s.db.Exec(ctx, `INSERT INTO signals(batch_id,code,severity,message,active,created_at) VALUES(?,?,?,?,?,?)`, signal.BatchID, signal.Code, signal.Severity, signal.Message, active, signal.CreatedAt)
	if err != nil {
		return model.Signal{}, fmt.Errorf("insert signal: %w", err)
	}
	signal.ID, err = txInsertID(result)
	if err != nil {
		return model.Signal{}, err
	}
	return signal, nil
}

func (s *SignalStore) ListActiveByBatch(ctx context.Context, batchID int64) ([]model.Signal, error) {
	rows, err := s.db.Query(ctx, `SELECT id,batch_id,code,severity,message,active,created_at FROM signals WHERE batch_id=? AND active=1 ORDER BY id DESC`, batchID)
	if err != nil {
		return nil, fmt.Errorf("list active signals: %w", err)
	}
	values := make([]model.Signal, 0)
	for rows.Next() {
		var signal model.Signal
		var active int
		if err := rows.Scan(&signal.ID, &signal.BatchID, &signal.Code, &signal.Severity, &signal.Message, &active, &signal.CreatedAt); err != nil {
			return nil, closeRows(rows, fmt.Errorf("scan signal: %w", err))
		}
		signal.Active = active == 1
		values = append(values, signal)
	}
	return values, closeRows(rows, nil)
}

func (s *SignalStore) Resolve(ctx context.Context, id int64) error {
	result, err := s.db.Exec(ctx, `UPDATE signals SET active=0 WHERE id=? AND active=1`, id)
	if err != nil {
		return fmt.Errorf("resolve signal: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read signal update: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SignalStore) CountBlockers(ctx context.Context, batchID int64) (int, error) {
	var count int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM signals WHERE batch_id=? AND active=1 AND severity=?`, batchID, model.SeverityBlocker).Scan(&count); err != nil {
		return 0, fmt.Errorf("count blockers: %w", err)
	}
	return count, nil
}
