package engine

import (
	"context"
	"sync"
	"time"
)

type Lifecycle struct {
	mu      sync.Mutex
	running bool
	started time.Time
	stop    chan struct{}
	done    chan struct{}
}

func NewLifecycle() *Lifecycle { return &Lifecycle{} }

func (l *Lifecycle) Start(interval time.Duration, fn func(context.Context)) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.running {
		return false
	}
	if interval <= 0 {
		interval = time.Minute
	}
	l.running = true
	l.started = time.Now().UTC()
	l.stop = make(chan struct{})
	l.done = make(chan struct{})
	stop, done := l.stop, l.done
	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				fn(ctx)
			}
		}
	}()
	return true
}

func (l *Lifecycle) Stop() {
	l.mu.Lock()
	if !l.running {
		l.mu.Unlock()
		return
	}
	stop, done := l.stop, l.done
	l.running = false
	l.mu.Unlock()
	close(stop)
	<-done
}

func (l *Lifecycle) Running() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.running
}

func (l *Lifecycle) StartedAt() time.Time {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.started
}

func (l *Lifecycle) RunOnce(ctx context.Context, fn func(context.Context)) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	fn(ctx)
	return ctx.Err()
}
