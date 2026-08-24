package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/jb843051627/moth-index/internal/model"
	"github.com/jb843051627/moth-index/internal/store"
)

type ConsistencyIssue struct {
	Code     string `json:"code"`
	Entity   string `json:"entity"`
	EntityID int64  `json:"entity_id"`
	Detail   string `json:"detail"`
}

type ConsistencyService struct {
	stations  *store.StationStore
	traps     *store.TrapStore
	batches   *store.BatchStore
	specimens *store.SpecimenStore
	readings  *store.ReadingStore
	reports   *store.ReportStore
}

func NewConsistencyService(stations *store.StationStore, traps *store.TrapStore, batches *store.BatchStore, specimens *store.SpecimenStore, readings *store.ReadingStore, reports *store.ReportStore) *ConsistencyService {
	return &ConsistencyService{stations: stations, traps: traps, batches: batches, specimens: specimens, readings: readings, reports: reports}
}

func (s *ConsistencyService) Batch(ctx context.Context, batchID int64) ([]ConsistencyIssue, error) {
	batch, err := s.batches.Get(ctx, batchID)
	if err != nil {
		return nil, err
	}
	issues := make([]ConsistencyIssue, 0)
	station, err := s.stations.Get(ctx, batch.StationID)
	if err != nil {
		return nil, err
	}
	if station.Status == model.StationRetired && batch.Status != model.BatchArchived {
		issues = append(issues, ConsistencyIssue{Code: "retired-station", Entity: "batch", EntityID: batchID, Detail: "non-archived batch points to retired station"})
	}
	trap, err := s.traps.Get(ctx, batch.TrapID)
	if err != nil {
		return nil, err
	}
	if trap.StationID != batch.StationID {
		issues = append(issues, ConsistencyIssue{Code: "trap-station", Entity: "batch", EntityID: batchID, Detail: "trap belongs to another station"})
	}
	specimens, err := s.specimens.ListByBatch(ctx, batchID, 1000, 0)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{})
	for _, specimen := range specimens {
		tag := strings.TrimSpace(specimen.Tag)
		if _, exists := seen[tag]; exists {
			issues = append(issues, ConsistencyIssue{Code: "duplicate-tag", Entity: "specimen", EntityID: specimen.ID, Detail: "tag is repeated in batch"})
		}
		seen[tag] = struct{}{}
		if specimen.Status == model.SpecimenAccepted && !specimen.IsClassified() {
			issues = append(issues, ConsistencyIssue{Code: "accepted-unclassified", Entity: "specimen", EntityID: specimen.ID, Detail: "accepted specimen has no taxonomy"})
		}
	}
	readings, err := s.readings.ListByBatch(ctx, batchID, "", "", 1000)
	if err != nil {
		return nil, err
	}
	for _, reading := range readings {
		if reading.BatchID != batchID {
			issues = append(issues, ConsistencyIssue{Code: "reading-batch", Entity: "reading", EntityID: reading.ID, Detail: "reading points to another batch"})
		}
	}
	return issues, nil
}

func (s *ConsistencyService) Healthy(ctx context.Context, batchID int64) (bool, error) {
	issues, err := s.Batch(ctx, batchID)
	if err != nil {
		return false, err
	}
	return len(issues) == 0, nil
}

func (s *ConsistencyService) SummaryMatches(ctx context.Context, batchID int64) (bool, error) {
	summary, err := s.reports.BatchSummary(ctx, batchID)
	if err != nil {
		return false, err
	}
	count, err := s.specimens.CountByBatch(ctx, batchID)
	if err != nil {
		return false, err
	}
	if summary.Specimens != count {
		return false, fmt.Errorf("summary count mismatch: %d != %d", summary.Specimens, count)
	}
	return true, nil
}

func (s *ConsistencyService) Explain(issues []ConsistencyIssue) []string {
	output := make([]string, 0, len(issues))
	for _, issue := range issues {
		output = append(output, issue.Code+": "+issue.Detail)
	}
	return output
}
