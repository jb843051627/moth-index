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

func setupBug02(t *testing.T) (context.Context, *store.DB, *service.Registry, model.NightBatch) {
	ctx := context.Background()
	db, err := store.Open(filepath.Join(t.TempDir(), "case.db"))
	if err != nil {
		t.Fatal(err)
	}
	registry := service.NewRegistry(db)
	station, err := registry.Stations.Create(ctx, model.Station{Code: "BUG02", Name: "Night transect 02", Habitat: "forest edge", Timezone: "Asia/Shanghai"})
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	trap, err := registry.Traps.Register(ctx, model.Trap{StationID: station.ID, Label: "lamp-02", Kind: "uv-led"})
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


func TestBug02_CancelStopsBatchIngest(t *testing.T) {
	ctx, db, registry, batch := setupBug02(t)
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	values := []model.Reading{{BatchID: batch.ID, ObservedAt: time.Now().UTC().Format(time.RFC3339Nano), Temperature: 22, Humidity: 70, Lux: 12, Rainfall: 0, Source: "field"}}
	if _, err := registry.Readings.BatchIngest(ctx, values); err == nil {
		t.Fatal("cancelled batch unexpectedly succeeded")
	}
	var count int
	if err := db.QueryRow(context.Background(), "SELECT COUNT(*) FROM readings WHERE batch_id=?", batch.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("cancelled request wrote %d readings", count)
	}
}
