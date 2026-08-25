package regression

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/jb843051627/moth-index/internal/engine"
	"github.com/jb843051627/moth-index/internal/model"
	"github.com/jb843051627/moth-index/internal/store"
)

type bug24Runner struct{}

func (bug24Runner) Process(context.Context, model.ReviewTask) error { return errors.New("review failed") }

func TestBug24_FailedWorkerLeavesFailedTask(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "case.db"))
	if err != nil { t.Fatal(err) }
	defer db.Close()
	ctx := context.Background()
	task, err := store.NewTaskStore(db).Enqueue(ctx, model.ReviewTask{Kind: "specimen", RefID: 1, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano)})
	if err != nil { t.Fatal(err) }
	queue := engine.NewQueue(store.NewTaskStore(db), bug24Runner{}, 1)
	if err := queue.Submit(ctx, task.ID); err != nil { t.Fatal(err) }
	for i := 0; i < 20; i++ {
		var status string
		if err := db.QueryRow(ctx, "SELECT status FROM review_tasks WHERE id=?", task.ID).Scan(&status); err != nil { t.Fatal(err) }
		if status != model.TaskQueued && status != model.TaskRunning { break }
		time.Sleep(5 * time.Millisecond)
	}
	queue.Close()
	var status string
	if err := db.QueryRow(ctx, "SELECT status FROM review_tasks WHERE id=?", task.ID).Scan(&status); err != nil { t.Fatal(err) }
	if status != model.TaskFailed { t.Fatalf("task status=%s", status) }
}
