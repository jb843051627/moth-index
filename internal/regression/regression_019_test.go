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

func setupBug19(t *testing.T) (context.Context, *store.DB, *service.Registry, model.NightBatch) {
	ctx := context.Background()
	db, err := store.Open(filepath.Join(t.TempDir(), "case.db"))
	if err != nil {
		t.Fatal(err)
	}
	registry := service.NewRegistry(db)
	station, err := registry.Stations.Create(ctx, model.Station{Code: "BUG19", Name: "Night transect 19", Habitat: "forest edge", Timezone: "Asia/Shanghai"})
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	trap, err := registry.Traps.Register(ctx, model.Trap{StationID: station.ID, Label: "lamp-19", Kind: "uv-led"})
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


func TestBug19_DailyReportDoesNotMultiplyRows(t *testing.T) {
	ctx, _, registry, batch := setupBug19(t)
	for _, tag := range []string{"M-19-A", "M-19-B"} {
		if _, err := registry.Specimens.Capture(ctx, model.Specimen{BatchID: batch.ID, Tag: tag, Count: 1}); err != nil { t.Fatal(err) }
	}
	for _, stamp := range []string{"2026-08-24T01:00:00Z", "2026-08-24T02:00:00Z"} {
		if _, err := registry.Readings.Ingest(ctx, model.Reading{BatchID: batch.ID, ObservedAt: stamp, Temperature: 20, Humidity: 70, Lux: 10, Source: "field"}); err != nil { t.Fatal(err) }
	}
	records, err := registry.Reports.Daily(ctx, "", "", 10)
	if err != nil { t.Fatal(err) }
	if len(records) != 1 || records[0].Specimens != 2 || records[0].Captures != 2 { t.Fatalf("daily record=%v", records) }
}
