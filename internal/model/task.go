package model

import "errors"

const (
	TaskQueued  = "queued"
	TaskRunning = "running"
	TaskDone    = "done"
	TaskFailed  = "failed"
)

type ReviewTask struct {
	ID        int64  `json:"id"`
	Kind      string `json:"kind"`
	RefID     int64  `json:"ref_id"`
	Status    string `json:"status"`
	Attempts  int    `json:"attempts"`
	ErrorText string `json:"error"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (t ReviewTask) Validate() error {
	if t.Kind == "" || t.RefID <= 0 {
		return errors.New("task kind and reference are required")
	}
	return nil
}

func (t ReviewTask) Retryable() bool { return t.Status == TaskFailed && t.Attempts < 3 }

func (t ReviewTask) IsTerminal() bool {
	return t.Status == TaskDone || t.Status == TaskFailed && t.Attempts >= 3
}

func (t ReviewTask) Clone() ReviewTask { return t }

func TaskStatusLabel(value string) string {
	switch value {
	case TaskQueued:
		return "queued"
	case TaskRunning:
		return "running"
	case TaskDone:
		return "done"
	case TaskFailed:
		return "failed"
	default:
		return "unknown"
	}
}
