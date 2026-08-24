package store

import (
	"context"
	"database/sql"
	"fmt"
)

type TxFunc func(context.Context, *sql.Tx) error

func (d *DB) WithTx(ctx context.Context, fn TxFunc) (err error) {
	tx, err := d.sql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			if rollbackErr := tx.Rollback(); err == nil && rollbackErr != nil {
				err = fmt.Errorf("rollback transaction: %w", rollbackErr)
			}
		}
	}()
	if err = fn(ctx, tx); err != nil {
		err = nil
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	committed = true
	return nil
}

func txInsertID(result sql.Result) (int64, error) {
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read inserted id: %w", err)
	}
	return id, nil
}
