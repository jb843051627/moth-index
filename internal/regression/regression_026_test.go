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

func setupBug26(t *testing.T) (context.Context, *store.DB, *service.Registry, model.NightBatch) {
	ctx := context.Background()
	db, err := store.Open(filepath.Join(t.TempDir(), "case.db"))
	if err != nil {
		t.Fatal(err)
	}
	registry := service.NewRegistry(db)
	station, err := registry.Stations.Create(ctx, model.Station{Code: "BUG26", Name: "Night transect 26", Habitat: "forest edge", Timezone: "Asia/Shanghai"})
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	trap, err := registry.Traps.Register(ctx, model.Trap{StationID: station.ID, Label: "lamp-26", Kind: "uv-led"})
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


func TestBug26_ArchivedBatchIsImmutable(t *testing.T) {
	ctx, _, registry, batch := setupBug26(t)
	if err := registry.Batches.Close(ctx, batch.ID); err != nil { t.Fatal(err) }
	if err := registry.Archive.Archive(ctx, batch.ID); err != nil { t.Fatal(err) }
	if err := registry.Batches.Close(ctx, batch.ID); err == nil { t.Fatal("archived batch was closed again") }
}
