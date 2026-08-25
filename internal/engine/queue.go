package engine

import (
	"context"
	"errors"
	"sync"

	"github.com/jb843051627/moth-index/internal/model"
	"github.com/jb843051627/moth-index/internal/store"
)

type Runner interface {
	Process(context.Context, model.ReviewTask) error
}

type Queue struct {
	tasks   *store.TaskStore
	runner  Runner
	jobs    chan int64
	ctx     context.Context
	cancel  context.CancelFunc
	workers int
	done    chan struct{}
	wg      sync.WaitGroup
	once    sync.Once
	metrics *Metrics
}

func NewQueue(tasks *store.TaskStore, runner Runner, workers int) *Queue {
	if workers < 1 {
		workers = 1
	}
	ctx, cancel := context.WithCancel(context.Background())
	queue := &Queue{tasks: tasks, runner: runner, jobs: make(chan int64, workers*4), ctx: ctx, cancel: cancel, workers: workers, done: make(chan struct{}), metrics: NewMetrics()}
	queue.Start()
	return queue
}

func (q *Queue) Start() {
	q.wg.Add(q.workers)
	for i := 0; i < q.workers; i++ {
		go q.worker(i)
	}
}

func (q *Queue) Submit(ctx context.Context, taskID int64) error {
	if taskID <= 0 {
		return errors.New("task id is required")
	}
	select {
	case q.jobs <- taskID:
		q.metrics.Add("submitted", 1)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-q.ctx.Done():
		return errors.New("queue is closed")
	}
}

func (q *Queue) Close() {
	q.cancel()
	q.once.Do(func() { close(q.done) })
	q.wg.Wait()
}

func (q *Queue) Wait() {
	q.Close()
	q.wg.Wait()
}

func (q *Queue) Size() int { return len(q.jobs) }

func (q *Queue) Metrics() map[string]int64 { return q.metrics.Snapshot() }

func (q *Queue) worker(index int) {
	defer q.wg.Done()
	for {
		select {
		case <-q.ctx.Done():
			return
		case <-q.done:
			return
		case <-q.jobs:
			q.runOne()
		}
	}
}

func (q *Queue) runOne() {
	task, err := q.tasks.Claim(q.ctx)
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			q.metrics.Add("claim_errors", 1)
		}
		return
	}
	q.metrics.Add("claimed", 1)
	if err := q.runner.Process(q.ctx, task); err != nil {
		q.metrics.Add("failed", 1)
		_ = q.tasks.Fail(context.Background(), task.ID, err.Error())
		return
	}
	q.metrics.Add("completed", 1)
	_ = q.tasks.Complete(context.Background(), task.ID)
}
