package workerq

import (
	"context"
	"fmt"

	"go.uber.org/multierr"
)

type WorkerFunc func(cmd *Worker) error

type Worker struct {
	ctx    context.Context
	cancel context.CancelFunc
	work   WorkerFunc
	errs   error
}

func NewWorker(ctx context.Context, work WorkerFunc) *Worker {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	return &Worker{
		ctx:    ctx,
		cancel: cancel,
		work:   work,
	}
}

func (c *Worker) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *Worker) Wait() {
	<-c.ctx.Done()
}

func (c *Worker) Err() error {
	return c.errs
}

func (c *Worker) Do() {
	defer func() {
		if err := recover(); err != nil {
			c.errs = multierr.Append(c.errs, fmt.Errorf("panic: %v", err))
		}
		c.cancel() // mark work done
	}()
	if c.work != nil {
		err := c.work(c)
		if err != nil {
			c.errs = multierr.Append(c.errs, err)
		}
	}
}

type WorkerQueue struct {
	workers chan *Worker
}

func NewWorkerQueue(concurrency int) *WorkerQueue {
	if concurrency <= 0 {
		concurrency = 1
	}

	q := &WorkerQueue{
		workers: make(chan *Worker, concurrency),
	}
	return q
}

func (q *WorkerQueue) SetConcurrency(concurrency int) {
	if cap(q.workers) != concurrency {
		return
	}
	// TODO: resize workers
}

func (q *WorkerQueue) AppendWorker(worker *Worker) <-chan struct{} {
	go q.dispatchWorker(worker)
	return worker.Done()
}

func (q *WorkerQueue) AppendWorkerFunc(ctx context.Context, wf WorkerFunc) <-chan struct{} {
	return q.AppendWorker(NewWorker(ctx, wf))
}

func (q *WorkerQueue) AppendWorkerWait(worker *Worker) error {
	<-q.AppendWorker(worker)
	return worker.Err()
}

func (q *WorkerQueue) AppendWorkerFuncWait(ctx context.Context, wf WorkerFunc) error {
	return q.AppendWorkerWait(NewWorker(ctx, wf))
}

func (q *WorkerQueue) dispatchWorker(worker *Worker) {
	q.workers <- worker
	defer func() { <-q.workers }()
	worker.Do()
}
