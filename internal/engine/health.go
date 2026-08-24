package engine

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Check func(context.Context) error

type CheckResult struct {
	Name    string    `json:"name"`
	Status  string    `json:"status"`
	Detail  string    `json:"detail"`
	Checked time.Time `json:"checked"`
}

type HealthMonitor struct {
	mu      sync.RWMutex
	checks  map[string]Check
	results map[string]CheckResult
}

func NewHealthMonitor() *HealthMonitor {
	return &HealthMonitor{checks: make(map[string]Check), results: make(map[string]CheckResult)}
}

func (m *HealthMonitor) Register(name string, check Check) error {
	if name == "" || check == nil {
		return fmt.Errorf("health check needs name and function")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.checks[name]; exists {
		return fmt.Errorf("health check %s already exists", name)
	}
	m.checks[name] = check
	return nil
}

func (m *HealthMonitor) Remove(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.checks, name)
	delete(m.results, name)
}

func (m *HealthMonitor) Run(ctx context.Context) []CheckResult {
	m.mu.RLock()
	checks := make(map[string]Check, len(m.checks))
	for name, check := range m.checks {
		checks[name] = check
	}
	m.mu.RUnlock()
	output := make([]CheckResult, 0, len(checks))
	for name, check := range checks {
		result := CheckResult{Name: name, Status: "ok", Checked: time.Now().UTC()}
		if err := check(ctx); err != nil {
			result.Status = "failed"
			result.Detail = err.Error()
		}
		output = append(output, result)
		m.mu.Lock()
		m.results[name] = result
		m.mu.Unlock()
	}
	return output
}

func (m *HealthMonitor) Snapshot() []CheckResult {
	m.mu.RLock()
	defer m.mu.RUnlock()
	output := make([]CheckResult, 0, len(m.results))
	for _, result := range m.results {
		output = append(output, result)
	}
	return output
}

func (m *HealthMonitor) Healthy() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, result := range m.results {
		if result.Status != "ok" {
			return false
		}
	}
	return true
}

func (m *HealthMonitor) RunWithTimeout(parent context.Context, timeout time.Duration) []CheckResult {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	return m.Run(ctx)
}
