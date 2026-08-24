package engine

import (
	"context"
	"errors"
	"sync"
	"time"
)

type WorkerGroup struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewWorkerGroup(parent context.Context) *WorkerGroup {
	ctx, cancel := context.WithCancel(parent)
	return &WorkerGroup{ctx: ctx, cancel: cancel}
}

func (g *WorkerGroup) Go(fn func(context.Context) error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		_ = fn(g.ctx)
	}()
}

func (g *WorkerGroup) Stop() {
	g.cancel()
	g.wg.Wait()
}

func Retry(ctx context.Context, attempts int, delay time.Duration, fn func(context.Context) error) error {
	if attempts < 1 {
		attempts = 1
	}
	var last error
	for i := 0; i < attempts; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := fn(ctx); err == nil {
			return nil
		} else {
			last = err
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	if last == nil {
		return errors.New("retry failed")
	}
	return last
}
