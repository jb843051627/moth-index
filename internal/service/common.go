package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jb843051627/moth-index/internal/model"
	"github.com/jb843051627/moth-index/internal/store"
)

type Registry struct {
	DB          *store.DB
	Tasks       *store.TaskStore
	Stations    *StationService
	Traps       *TrapService
	Batches     *BatchService
	Specimens   *SpecimenService
	Readings    *ReadingService
	Taxonomy    *TaxonomyService
	Reviews     *ReviewService
	Quality     *QualityService
	Reports     *ReportService
	Archive     *ArchiveService
	Analysis    *AnalysisService
	Maintenance *MaintenanceService
	Phenology   *PhenologyService
	Seasonal    *SeasonalService
	Consistency *ConsistencyService
}

func NewRegistry(db *store.DB) *Registry {
	stations := store.NewStationStore(db)
	traps := store.NewTrapStore(db)
	batches := store.NewBatchStore(db)
	specimens := store.NewSpecimenStore(db)
	readings := store.NewReadingStore(db)
	taxonomy := store.NewTaxonomyStore(db)
	reviews := store.NewReviewStore(db)
	signals := store.NewSignalStore(db)
	tasks := store.NewTaskStore(db)
	reports := store.NewReportStore(db)
	maintenance := store.NewMaintenanceStore(db)
	events := store.NewEventStore(db)
	seasons := store.NewSeasonStore(db)
	return &Registry{
		DB:          db,
		Tasks:       tasks,
		Stations:    NewStationService(stations, traps, reports),
		Traps:       NewTrapService(traps, stations, batches),
		Batches:     NewBatchService(batches, stations, traps, reports),
		Specimens:   NewSpecimenService(specimens, batches, taxonomy, tasks, reports),
		Readings:    NewReadingService(readings, batches, reports),
		Taxonomy:    NewTaxonomyService(taxonomy, specimens),
		Reviews:     NewReviewService(reviews, specimens, reports, tasks),
		Quality:     NewQualityService(reports, readings, signals, batches),
		Reports:     NewReportService(reports, batches, readings, specimens),
		Archive:     NewArchiveService(batches, readings, signals, reports),
		Analysis:    NewAnalysisService(specimens, readings, batches, reports),
		Maintenance: NewMaintenanceService(maintenance, events),
		Phenology:   NewPhenologyService(batches, readings, specimens, traps),
		Seasonal:    NewSeasonalService(batches, readings, specimens, traps, seasons),
		Consistency: NewConsistencyService(stations, traps, batches, specimens, readings, reports),
	}
}

func serviceNow() string { return time.Now().UTC().Format(time.RFC3339Nano) }

func withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 8*time.Second)
}

func wrapValidation(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%v: %v", store.ErrValidation, err)
}

func isNotFound(err error) bool { return errors.Is(err, store.ErrNotFound) }

func requireStatus(actual string, allowed ...string) error {
	for _, candidate := range allowed {
		if actual == candidate {
			return nil
		}
	}
	return fmt.Errorf("%w: status %s", store.ErrState, actual)
}

func scoreReadings(values []model.Reading) float64 {
	if len(values) == 0 {
		return 0
	}
	total := 0.0
	for _, reading := range values {
		total += reading.QualityScore()
	}
	return total / float64(len(values))
}
