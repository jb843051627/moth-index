package regression

import (
	"context"
	"errors"
	"path/filepath"
	"database/sql"
	"testing"

	"github.com/jb843051627/moth-index/internal/store"
)

func TestBug03_TransactionReturnsCallbackError(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "case.db"))
	if err != nil { t.Fatal(err) }
	defer db.Close()
	sentinel := errors.New("review callback failed")
	err = db.WithTx(context.Background(), func(ctx context.Context, tx *sql.Tx) error {
		_, execErr := tx.ExecContext(ctx, "INSERT INTO audit_events(entity,entity_id,action,detail,created_at) VALUES(?,?,?,?,?)", "batch", 1, "review", "partial", "now")
		if execErr != nil { return execErr }
		return sentinel
	})
	if !errors.Is(err, sentinel) { t.Fatalf("WithTx error=%v", err) }
	var count int
	if err := db.QueryRow(context.Background(), "SELECT COUNT(*) FROM audit_events").Scan(&count); err != nil { t.Fatal(err) }
	if count != 0 { t.Fatalf("callback row committed: %d", count) }
}
