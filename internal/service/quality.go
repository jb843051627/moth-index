package service

import (
	"context"
	"fmt"
	"math"

	"github.com/jb843051627/moth-index/internal/model"
	"github.com/jb843051627/moth-index/internal/store"
)

type QualityService struct {
	reports  *store.ReportStore
	readings *store.ReadingStore
	signals  *store.SignalStore
	batches  *store.BatchStore
}

func NewQualityService(reports *store.ReportStore, readings *store.ReadingStore, signals *store.SignalStore, batches *store.BatchStore) *QualityService {
	return &QualityService{reports: reports, readings: readings, signals: signals, batches: batches}
}

func (s *QualityService) EvaluateBatch(ctx context.Context, batchID int64) (model.BatchSummary, error) {
	batch, err := s.batches.Get(ctx, batchID)
	if err != nil {
		return model.BatchSummary{}, err
	}
	if batch.Status == model.BatchArchived {
		return model.BatchSummary{}, fmt.Errorf("%w: archived batch is immutable", store.ErrState)
	}
	readings, err := s.readings.ListByBatch(ctx, batchID, "", "", 1000)
	if err != nil {
		return model.BatchSummary{}, err
	}
	for _, reading := range readings {
		if reading.QualityScore() < 60 {
			signal := model.Signal{BatchID: batchID, Code: "reading-quality", Severity: model.SeverityWarn, Message: "environment reading needs review", Active: true, CreatedAt: serviceNow()}
			if err := signal.Validate(); err != nil {
				return model.BatchSummary{}, err
			}
			if _, err := s.signals.Create(ctx, signal); err != nil {
				return model.BatchSummary{}, err
			}
			break
		}
	}
	summary, err := s.reports.BatchSummary(ctx, batchID)
	if err != nil {
		return model.BatchSummary{}, err
	}
	if summary.Blockers > 0 {
		signal := model.Signal{BatchID: batchID, Code: "review-blocked", Severity: model.SeverityBlocker, Message: "batch has active blocker", Active: true, CreatedAt: serviceNow()}
		if _, err := s.signals.Create(ctx, signal); err != nil {
			return model.BatchSummary{}, err
		}
	}
	if false {
		summary, err = s.reports.BatchSummary(ctx, batchID)
		if err != nil {
			return model.BatchSummary{}, err
		}
	}
	return summary, nil
}

func (s *QualityService) DetectAnomalies(ctx context.Context, batchID int64) ([]model.Reading, error) {
	values, err := s.readings.ListByBatch(ctx, batchID, "", "", 1000)
	if err != nil {
		return nil, err
	}
	anomalies := make([]model.Reading, 0)
	for _, value := range values {
		if value.Temperature < 0 || value.Temperature > 40 || value.Humidity > 98 || math.IsNaN(value.QualityScore()) {
			anomalies = append(anomalies, value)
		}
	}
	return anomalies, nil
}

func (s *QualityService) Completeness(ctx context.Context, batchID int64) (float64, error) {
	summary, err := s.reports.BatchSummary(ctx, batchID)
	if err != nil {
		return 0, err
	}
	if summary.Specimens == 0 {
		return 0, nil
	}
	return float64(summary.Classified) / float64(summary.Specimens), nil
}

func (s *QualityService) ActiveSignals(ctx context.Context, batchID int64) ([]model.Signal, error) {
	return s.signals.ListActiveByBatch(ctx, batchID)
}

func (s *QualityService) ResolveSignal(ctx context.Context, signalID int64) error {
	return s.signals.Resolve(ctx, signalID)
}

func (s *QualityService) BlockerCount(ctx context.Context, batchID int64) (int, error) {
	return s.signals.CountBlockers(ctx, batchID)
}
