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

func setupBug08(t *testing.T) (context.Context, *store.DB, *service.Registry, model.NightBatch) {
	ctx := context.Background()
	db, err := store.Open(filepath.Join(t.TempDir(), "case.db"))
	if err != nil {
		t.Fatal(err)
	}
	registry := service.NewRegistry(db)
	station, err := registry.Stations.Create(ctx, model.Station{Code: "BUG08", Name: "Night transect 08", Habitat: "forest edge", Timezone: "Asia/Shanghai"})
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	trap, err := registry.Traps.Register(ctx, model.Trap{StationID: station.ID, Label: "lamp-08", Kind: "uv-led"})
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


func TestBug08_BatchKeepsTrapOwnership(t *testing.T) {
	ctx, db, registry, batch := setupBug08(t)
	if err := registry.Batches.Close(ctx, batch.ID); err != nil { t.Fatal(err) }
	station, err := registry.Stations.Create(ctx, model.Station{Code: "OTHER08", Name: "Other station", Timezone: "Asia/Shanghai"})
	if err != nil { t.Fatal(err) }
	if _, err := registry.Batches.Open(ctx, model.NightBatch{StationID: station.ID, TrapID: batch.TrapID, StartedAt: time.Now().UTC().Format(time.RFC3339Nano)}); err == nil {
		t.Fatal("cross-station trap was accepted")
	}
	trap, err := registry.Traps.Get(ctx, batch.TrapID)
	if err != nil { t.Fatal(err) }
	if err := registry.Traps.Break(ctx, trap.ID); err != nil { t.Fatal(err) }
	if _, err := registry.Batches.Open(ctx, model.NightBatch{StationID: batch.StationID, TrapID: trap.ID, StartedAt: time.Now().UTC().Format(time.RFC3339Nano)}); err == nil {
		t.Fatal("broken trap was accepted")
	}
	_ = db
}
