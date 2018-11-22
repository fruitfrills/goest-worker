package goest_worker

import (
	"context"
)

type worker struct {

	// Chan for jobs
	task 				chan jobCall

	// Chan for send free message
	pool				WorkerPoolType

	// counter from pool for decrement waiting tasks
	counter 			Counter

	// worker context
	ctx 				context.Context

	// cancel context
	cancel				context.CancelFunc
}

func newWorker(ctx context.Context, workerPool WorkerPoolType, counter Counter) (WorkerInterface)  {
	ctx, cancel := context.WithCancel(ctx)
	return &worker{
		ctx: ctx,
		cancel: cancel,
		task: make(chan jobCall),
		pool: workerPool,
		counter: counter,
	}
}

func (w *worker) Start() {
	go func() {
		for {
			select {
			case <- w.ctx.Done():
				return
			case w.pool <- w:
				break
			}

			select {
			case <- w.ctx.Done():
				return
			case job := <- w.task:
				if job == nil {
					return
				}
				job.call()
				if w.counter != nil {
					w.counter.Decrement()
				}
			}
		}
	}()
}

func (w *worker) Context() context.Context  {
	return w.ctx
}

func (w *worker) Cancel()  {
	w.cancel()
}

func (w *worker) AddJob(job jobCall) {
	w.task <- job
	return
}