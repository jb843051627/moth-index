package engine

import (
	"sort"
	"sync"
	"time"
)

type Sample struct {
	Name  string    `json:"name"`
	Value float64   `json:"value"`
	At    time.Time `json:"at"`
}

type Telemetry struct {
	mu      sync.RWMutex
	samples map[string][]Sample
	limit   int
}

func NewTelemetry(limit int) *Telemetry {
	if limit < 20 {
		limit = 200
	}
	return &Telemetry{samples: make(map[string][]Sample), limit: limit}
}

func (t *Telemetry) Record(name string, value float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	values := append(t.samples[name], Sample{Name: name, Value: value, At: time.Now().UTC()})
	if len(values) > t.limit {
		values = values[len(values)-t.limit:]
	}
	t.samples[name] = values
}

func (t *Telemetry) Values(name string) []Sample {
	t.mu.RLock()
	defer t.mu.RUnlock()
	values := append([]Sample(nil), t.samples[name]...)
	return values
}

func (t *Telemetry) Latest(name string) (Sample, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	values := t.samples[name]
	if len(values) == 0 {
		return Sample{}, false
	}
	return values[len(values)-1], true
}

func (t *Telemetry) Average(name string, since time.Time) float64 {
	values := t.Values(name)
	total := 0.0
	count := 0
	for _, value := range values {
		if !value.At.Before(since) {
			total += value.Value
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

func (t *Telemetry) Names() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	names := make([]string, 0, len(t.samples))
	for name := range t.samples {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (t *Telemetry) Prune(since time.Time) int {
	t.mu.Lock()
	defer t.mu.Unlock()
	removed := 0
	for name, values := range t.samples {
		kept := values[:0]
		for _, value := range values {
			if value.At.Before(since) {
				removed++
				continue
			}
			kept = append(kept, value)
		}
		t.samples[name] = append([]Sample(nil), kept...)
	}
	return removed
}

func (t *Telemetry) Size() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	total := 0
	for _, values := range t.samples {
		total += len(values)
	}
	return total
}
