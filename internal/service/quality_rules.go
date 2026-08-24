package service

import (
	"context"
	"fmt"
	"math"

	"github.com/jb843051627/moth-index/internal/model"
	"github.com/jb843051627/moth-index/internal/store"
)

type RuleResult struct {
	Name     string  `json:"name"`
	Passed   bool    `json:"passed"`
	Score    float64 `json:"score"`
	Evidence string  `json:"evidence"`
}

type QualityRulesService struct {
	readings  *store.ReadingStore
	specimens *store.SpecimenStore
	batches   *store.BatchStore
}

func NewQualityRulesService(readings *store.ReadingStore, specimens *store.SpecimenStore, batches *store.BatchStore) *QualityRulesService {
	return &QualityRulesService{readings: readings, specimens: specimens, batches: batches}
}

func (s *QualityRulesService) Run(ctx context.Context, batchID int64) ([]RuleResult, error) {
	batch, err := s.batches.Get(ctx, batchID)
	if err != nil {
		return nil, err
	}
	readings, err := s.readings.ListByBatch(ctx, batchID, "", "", 1000)
	if err != nil {
		return nil, err
	}
	specimens, err := s.specimens.ListByBatch(ctx, batchID, 1000, 0)
	if err != nil {
		return nil, err
	}
	result := []RuleResult{
		s.readingCoverage(readings),
		s.weatherStability(readings),
		s.taxonomyCoverage(specimens),
		s.nightBoundary(batch, readings),
	}
	return result, nil
}

func (s *QualityRulesService) readingCoverage(values []model.Reading) RuleResult {
	if len(values) == 0 {
		return RuleResult{Name: "reading-coverage", Evidence: "no readings"}
	}
	score := 1.0
	if len(values) < 3 {
		score = float64(len(values)) / 3
	}
	return RuleResult{Name: "reading-coverage", Passed: score >= 1, Score: score, Evidence: fmt.Sprintf("%d readings", len(values))}
}

func (s *QualityRulesService) weatherStability(values []model.Reading) RuleResult {
	if len(values) < 2 {
		return RuleResult{Name: "weather-stability", Score: 0, Evidence: "not enough readings"}
	}
	variance := model.ReadingVariance(values)
	score := 1 / (1 + variance)
	if math.IsNaN(score) || math.IsInf(score, 0) {
		score = 0
	}
	return RuleResult{Name: "weather-stability", Passed: score >= 0.5, Score: score, Evidence: fmt.Sprintf("temperature variance %.3f", variance)}
}

func (s *QualityRulesService) taxonomyCoverage(values []model.Specimen) RuleResult {
	if len(values) == 0 {
		return RuleResult{Name: "taxonomy-coverage", Passed: true, Score: 1, Evidence: "no specimens"}
	}
	classified := 0
	for _, value := range values {
		if value.IsClassified() {
			classified++
		}
	}
	score := float64(classified) / float64(len(values))
	return RuleResult{Name: "taxonomy-coverage", Passed: score >= 0.8, Score: score, Evidence: fmt.Sprintf("%d/%d classified", classified, len(values))}
}

func (s *QualityRulesService) nightBoundary(batch model.NightBatch, values []model.Reading) RuleResult {
	if batch.StartedAt == "" || len(values) == 0 {
		return RuleResult{Name: "night-boundary", Evidence: "missing time context"}
	}
	start, err := model.ParseWindow(batch.StartedAt, values[len(values)-1].ObservedAt)
	if err != nil {
		return RuleResult{Name: "night-boundary", Evidence: "invalid timestamp"}
	}
	passed := start.Duration() >= 0
	return RuleResult{Name: "night-boundary", Passed: passed, Score: 1, Evidence: start.Duration().String()}
}
