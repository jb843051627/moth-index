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

func setupBug25(t *testing.T) (context.Context, *store.DB, *service.Registry, model.NightBatch) {
	ctx := context.Background()
	db, err := store.Open(filepath.Join(t.TempDir(), "case.db"))
	if err != nil {
		t.Fatal(err)
	}
	registry := service.NewRegistry(db)
	station, err := registry.Stations.Create(ctx, model.Station{Code: "BUG25", Name: "Night transect 25", Habitat: "forest edge", Timezone: "Asia/Shanghai"})
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	trap, err := registry.Traps.Register(ctx, model.Trap{StationID: station.ID, Label: "lamp-25", Kind: "uv-led"})
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


func TestBug25_AcceptedSpecimenCannotBeRejected(t *testing.T) {
	ctx, _, registry, batch := setupBug25(t)
	specimen, err := registry.Specimens.Capture(ctx, model.Specimen{BatchID: batch.ID, Tag: "M-25", Count: 1})
	if err != nil { t.Fatal(err) }
	if _, err := registry.Specimens.Classify(ctx, specimen.ID, model.Taxon{Family: "Noctuidae", Genus: "Acronicta", Species: "rumicis"}); err != nil { t.Fatal(err) }
	if err := registry.Specimens.Accept(ctx, specimen.ID); err != nil { t.Fatal(err) }
	if err := registry.Specimens.Reject(ctx, specimen.ID, "late note"); err == nil { t.Fatal("accepted specimen was rejected") }
	current, err := registry.Specimens.Get(ctx, specimen.ID)
	if err != nil { t.Fatal(err) }
	if current.Status != model.SpecimenAccepted { t.Fatalf("status=%s", current.Status) }
}
