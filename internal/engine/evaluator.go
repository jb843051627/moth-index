package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/jb843051627/moth-index/internal/model"
)

type Evaluation struct {
	BatchID    int64
	Readings   int
	WarmNights int
	WetNights  int
	Score      float64
	Generated  string
}

type Evaluator struct {
	cache   *ReadingCache
	metrics *Metrics
}

func NewEvaluator(cache *ReadingCache, metrics *Metrics) *Evaluator {
	return &Evaluator{cache: cache, metrics: metrics}
}

func (e *Evaluator) Evaluate(ctx context.Context, batchID int64) (Evaluation, error) {
	if err := ctx.Err(); err != nil {
		return Evaluation{}, err
	}
	values := e.cache.Get(batchID)
	if len(values) == 0 {
		return Evaluation{}, fmt.Errorf("no cached readings for batch %d", batchID)
	}
	result := Evaluation{BatchID: batchID, Readings: len(values), Generated: time.Now().UTC().Format(time.RFC3339Nano)}
	total := 0.0
	for _, reading := range values {
		if err := ctx.Err(); err != nil {
			return Evaluation{}, err
		}
		if reading.IsWarm() {
			result.WarmNights++
		}
		if reading.IsWet() {
			result.WetNights++
		}
		total += reading.QualityScore()
	}
	result.Score = total / float64(len(values))
	e.metrics.Add("evaluated", 1)
	return result, nil
}

func (e *Evaluator) Observe(reading model.Reading) { e.cache.Put(reading.BatchID, reading) }

func (e *Evaluator) CacheSize(batchID int64) int { return len(e.cache.Get(batchID)) }
