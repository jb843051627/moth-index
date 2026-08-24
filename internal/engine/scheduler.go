package engine

import (
	"context"
	"time"
)

type Scheduler struct {
	group *WorkerGroup
}

func NewScheduler(parent context.Context) *Scheduler {
	return &Scheduler{group: NewWorkerGroup(parent)}
}

func (s *Scheduler) Every(interval time.Duration, fn func(context.Context) error) {
	if interval <= 0 {
		interval = time.Minute
	}
	s.group.Go(func(ctx context.Context) error {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				if err := fn(ctx); err != nil {
					continue
				}
			}
		}
	})
}

func (s *Scheduler) Stop() { s.group.Stop() }
