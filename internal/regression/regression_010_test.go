package regression

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/jb843051627/moth-index/internal/model"
	"github.com/jb843051627/moth-index/internal/store"
)

func TestBug10_ClaimDoesNotReclaimRunningTask(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "case.db"))
	if err != nil { t.Fatal(err) }
	defer db.Close()
	ctx := context.Background()
	tasks := store.NewTaskStore(db)
	task, err := tasks.Enqueue(ctx, model.ReviewTask{Kind: "specimen", RefID: 1, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano)})
	if err != nil { t.Fatal(err) }
	if _, err := tasks.Claim(ctx); err != nil { t.Fatal(err) }
	if _, err := tasks.Claim(ctx); !errors.Is(err, store.ErrNotFound) { t.Fatalf("second claim error=%v for task %d", err, task.ID) }
}
