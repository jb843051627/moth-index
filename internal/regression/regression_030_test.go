package regression

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/jb843051627/moth-index/internal/model"
	"github.com/jb843051627/moth-index/internal/service"
	"github.com/jb843051627/moth-index/internal/store"
)

var _ = errors.Is

func setupBug30(t *testing.T) (context.Context, *store.DB, *service.Registry, model.NightBatch) {
	ctx := context.Background()
	db, err := store.Open(filepath.Join(t.TempDir(), "case.db"))
	if err != nil {
		t.Fatal(err)
	}
	registry := service.NewRegistry(db)
	station, err := registry.Stations.Create(ctx, model.Station{Code: "BUG30", Name: "Night transect 30", Habitat: "forest edge", Timezone: "Asia/Shanghai"})
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	trap, err := registry.Traps.Register(ctx, model.Trap{StationID: station.ID, Label: "lamp-30", Kind: "uv-led"})
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	batch, err := registry.Batches.Open(ctx, model.NightBatch{StationID: station.ID, TrapID: trap.ID, StartedAt: time.Now().UTC().Add(-2 * time.Hour).Format(time.RFC3339Nano)})
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return ctx, db, registry, batch
}


func TestBug30_EvaluationReturnsFreshBlockerSummary(t *testing.T) {
	ctx, db, registry, batch := setupBug30(t)
	_, err := db.Exec(ctx, "INSERT INTO signals(batch_id,code,severity,message,active,created_at) VALUES(?,?,?,?,?,?)", batch.ID, "existing", model.SeverityBlocker, "unreviewed", 1, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil { t.Fatal(err) }
	summary, err := registry.Quality.EvaluateBatch(ctx, batch.ID)
	if err != nil { t.Fatal(err) }
	if summary.Blockers != 2 { t.Fatalf("blockers=%d, want 2", summary.Blockers) }
}
