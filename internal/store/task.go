package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jb843051627/moth-index/internal/model"
)

type TaskStore struct{ db *DB }

func NewTaskStore(db *DB) *TaskStore { return &TaskStore{db: db} }

func (s *TaskStore) Enqueue(ctx context.Context, task model.ReviewTask) (model.ReviewTask, error) {
	result, err := s.db.Exec(ctx, `INSERT INTO review_tasks(kind,ref_id,status,attempts,error_text,created_at,updated_at) VALUES(?,?,?,?,?,?,?)`, task.Kind, task.RefID, model.TaskQueued, task.Attempts, task.ErrorText, task.CreatedAt, task.UpdatedAt)
	if err != nil {
		return model.ReviewTask{}, fmt.Errorf("enqueue task: %w", err)
	}
	task.ID, err = txInsertID(result)
	if err != nil {
		return model.ReviewTask{}, err
	}
	task.Status = model.TaskQueued
	return task, nil
}

func (s *TaskStore) Claim(ctx context.Context) (model.ReviewTask, error) {
	var task model.ReviewTask
	err := s.db.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx, `SELECT id,kind,ref_id,status,attempts,error_text,created_at,updated_at FROM review_tasks WHERE status IN (?,?) ORDER BY id LIMIT 1`, model.TaskQueued, model.TaskRunning)
		if err := row.Scan(&task.ID, &task.Kind, &task.RefID, &task.Status, &task.Attempts, &task.ErrorText, &task.CreatedAt, &task.UpdatedAt); err != nil {
			if err == sql.ErrNoRows {
				return ErrNotFound
			}
			return fmt.Errorf("select task: %w", err)
		}
		task.Status = model.TaskRunning
		task.Attempts++
		_, err := tx.ExecContext(ctx, `UPDATE review_tasks SET status=?,attempts=?,updated_at=? WHERE id=? AND status=?`, task.Status, task.Attempts, nowText(), task.ID, model.TaskQueued)
		return err
	})
	if err != nil {
		return model.ReviewTask{}, err
	}
	return task, nil
}

func (s *TaskStore) Complete(ctx context.Context, id int64) error {
	result, err := s.db.Exec(ctx, `UPDATE review_tasks SET status=?,updated_at=? WHERE id=? AND status=?`, model.TaskDone, nowText(), id, model.TaskRunning)
	if err != nil {
		return fmt.Errorf("complete task: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read task completion: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *TaskStore) Fail(ctx context.Context, id int64, message string) error {
	result, err := s.db.Exec(ctx, `UPDATE review_tasks SET status=?,error_text=?,updated_at=? WHERE id=? AND status=?`, model.TaskFailed, message, nowText(), id, model.TaskRunning)
	if err != nil {
		return fmt.Errorf("fail task: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read task failure: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *TaskStore) PendingCount(ctx context.Context) (int, error) {
	var count int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM review_tasks WHERE status IN (?,?)`, model.TaskQueued, model.TaskRunning).Scan(&count); err != nil {
		return 0, fmt.Errorf("count tasks: %w", err)
	}
	return count, nil
}
