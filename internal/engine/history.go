package engine

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type Event struct {
	ID        int64     `json:"id"`
	Kind      string    `json:"kind"`
	Reference int64     `json:"reference"`
	Detail    string    `json:"detail"`
	At        time.Time `json:"at"`
}

type History struct {
	mu     sync.RWMutex
	next   int64
	events []Event
	limit  int
}

func NewHistory(limit int) *History {
	if limit < 10 {
		limit = 100
	}
	return &History{limit: limit, events: make([]Event, 0, limit)}
}

func (h *History) Add(kind string, reference int64, detail string) Event {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.next++
	event := Event{ID: h.next, Kind: kind, Reference: reference, Detail: detail, At: time.Now().UTC()}
	h.events = append(h.events, event)
	if len(h.events) > h.limit {
		h.events = append([]Event(nil), h.events[len(h.events)-h.limit:]...)
	}
	return event
}

func (h *History) List(kind string, reference int64) []Event {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make([]Event, 0)
	for _, event := range h.events {
		if kind != "" && event.Kind != kind {
			continue
		}
		if reference > 0 && event.Reference != reference {
			continue
		}
		result = append(result, event)
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].At.Before(result[j].At) })
	return result
}

func (h *History) Since(cutoff time.Time) []Event {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make([]Event, 0)
	for _, event := range h.events {
		if !event.At.Before(cutoff) {
			result = append(result, event)
		}
	}
	return result
}

func (h *History) Size() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.events)
}

func (h *History) Last() (Event, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.events) == 0 {
		return Event{}, fmt.Errorf("history is empty")
	}
	return h.events[len(h.events)-1], nil
}
