package engine

import "sync"

type Metrics struct {
	mu     sync.RWMutex
	values map[string]int64
}

func NewMetrics() *Metrics { return &Metrics{values: make(map[string]int64)} }

func (m *Metrics) Add(name string, amount int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.values[name] += amount
}

func (m *Metrics) Set(name string, value int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.values[name] = value
}

func (m *Metrics) Get(name string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.values[name]
}

func (m *Metrics) Snapshot() map[string]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	copyValues := make(map[string]int64, len(m.values))
	for key, value := range m.values {
		copyValues[key] = value
	}
	return copyValues
}

func (m *Metrics) Names() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	names := make([]string, 0, len(m.values))
	for key := range m.values {
		names = append(names, key)
	}
	return names
}
