package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jb843051627/moth-index/internal/model"
	"github.com/jb843051627/moth-index/internal/store"
)

type SurveyAnalysis struct {
	BatchID      int64                   `json:"batch_id"`
	Profile      model.CaptureProfile    `json:"profile"`
	Bands        []model.ObservationBand `json:"bands"`
	MedianTemp   float64                 `json:"median_temperature"`
	TempVariance float64                 `json:"temperature_variance"`
	ReadingScore float64                 `json:"reading_score"`
	Findings     []model.QualityFinding  `json:"findings"`
	WindowHours  float64                 `json:"window_hours"`
}

type AnalysisService struct {
	specimens *store.SpecimenStore
	readings  *store.ReadingStore
	batches   *store.BatchStore
	reports   *store.ReportStore
}

func NewAnalysisService(specimens *store.SpecimenStore, readings *store.ReadingStore, batches *store.BatchStore, reports *store.ReportStore) *AnalysisService {
	return &AnalysisService{specimens: specimens, readings: readings, batches: batches, reports: reports}
}

func (s *AnalysisService) Analyze(ctx context.Context, batchID int64) (SurveyAnalysis, error) {
	batch, err := s.batches.Get(ctx, batchID)
	if err != nil {
		return SurveyAnalysis{}, err
	}
	specimens, err := s.specimens.ListByBatch(ctx, batchID, 1000, 0)
	if err != nil {
		return SurveyAnalysis{}, err
	}
	readings, err := s.readings.ListByBatch(ctx, batchID, "", "", 1000)
	if err != nil {
		return SurveyAnalysis{}, err
	}
	output := SurveyAnalysis{BatchID: batchID, Profile: model.BuildCaptureProfile(specimens), Bands: model.RankBands(readings), MedianTemp: model.MedianTemperature(readings), TempVariance: model.ReadingVariance(readings), Findings: append(model.FindingsForReadings(readings), model.FindingsForSpecimens(specimens)...)}
	output.ReadingScore = scoreReadings(readings)
	if batch.EndedAt != "" {
		window, parseErr := model.ParseWindow(batch.StartedAt, batch.EndedAt)
		if parseErr != nil {
			return SurveyAnalysis{}, fmt.Errorf("parse batch window: %w", parseErr)
		}
		output.WindowHours = window.NightHours()
	} else {
		start, parseErr := time.Parse(time.RFC3339Nano, batch.StartedAt)
		if parseErr == nil {
			output.WindowHours = time.Since(start).Hours()
		}
	}
	return output, nil
}

func (s *AnalysisService) Profile(ctx context.Context, batchID int64) (model.CaptureProfile, error) {
	values, err := s.specimens.ListByBatch(ctx, batchID, 1000, 0)
	if err != nil {
		return model.CaptureProfile{}, err
	}
	return model.BuildCaptureProfile(values), nil
}

func (s *AnalysisService) Findings(ctx context.Context, batchID int64) ([]model.QualityFinding, error) {
	specimens, err := s.specimens.ListByBatch(ctx, batchID, 1000, 0)
	if err != nil {
		return nil, err
	}
	readings, err := s.readings.ListByBatch(ctx, batchID, "", "", 1000)
	if err != nil {
		return nil, err
	}
	findings := model.FindingsForReadings(readings)
	return append(findings, model.FindingsForSpecimens(specimens)...), nil
}

func (s *AnalysisService) NightWindow(ctx context.Context, batchID int64) (model.Window, error) {
	batch, err := s.batches.Get(ctx, batchID)
	if err != nil {
		return model.Window{}, err
	}
	if batch.EndedAt == "" {
		return model.Window{}, fmt.Errorf("%w: batch is still open", store.ErrState)
	}
	return model.ParseWindow(batch.StartedAt, batch.EndedAt)
}

func (s *AnalysisService) ReportQuality(ctx context.Context, batchID int64) (float64, error) {
	summary, err := s.reports.BatchSummary(ctx, batchID)
	if err != nil {
		return 0, err
	}
	if summary.LastReading == nil {
		return summary.Quality, nil
	}
	return summary.Quality * summary.LastReading.QualityScore() / 100, nil
}
