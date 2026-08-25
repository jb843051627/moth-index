package engine

import (
	"sync"
	"time"

	"github.com/jb843051627/moth-index/internal/model"
)

type ReadingCache struct {
	mu     sync.RWMutex
	values map[int64][]model.Reading
	order  []int64
}

func NewReadingCache() *ReadingCache { return &ReadingCache{values: make(map[int64][]model.Reading)} }

func (c *ReadingCache) Put(batchID int64, reading model.Reading) {
	c.mu.Lock()
	defer c.mu.Unlock()
	values := append(c.values[batchID], reading.Clone())
	c.values[batchID] = values
	if len(values) == 1 {
		c.order = append(c.order, batchID)
	}
}

func (c *ReadingCache) Get(batchID int64) []model.Reading {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.values[batchID]
}

func (c *ReadingCache) All() map[int64][]model.Reading {
	c.mu.RLock()
	defer c.mu.RUnlock()
	output := make(map[int64][]model.Reading, len(c.values))
	for id, values := range c.values {
		output[id] = values
	}
	return output
}

func (c *ReadingCache) PruneBefore(cutoff time.Time) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	removed := 0
	for batchID, values := range c.values {
		kept := values[:0]
		for _, reading := range values {
			stamp, err := time.Parse(time.RFC3339Nano, reading.ObservedAt)
			if err != nil || !stamp.Before(cutoff) {
				kept = append(kept, reading)
				continue
			}
			removed++
		}
		c.values[batchID] = append([]model.Reading(nil), kept...)
	}
	return removed
}

func cloneReadings(values []model.Reading) []model.Reading {
	output := make([]model.Reading, len(values))
	copy(output, values)
	return output
}
