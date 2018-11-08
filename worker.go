package goest_worker

import (
	"context"
	"sync/atomic"
)

type worker struct {

	// Chan for jobs
	task 				chan jobCall

	// Chan for send free message
	pool				WorkerPoolType

	// counter from pool for decrement waiting tasks
	counter 					*uint64
}

func newWorker(workerPool WorkerPoolType, counter *uint64) (WorkerInterface)  {
	return &worker{
		task: make(chan jobCall),
		pool: workerPool,
		counter: counter,
	}
}

func (w *worker) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <- ctx.Done():
				return
			default:
				w.pool <- w
			}

			select {
			case <- ctx.Done():
				return
			case job := <- w.task:
				if job == nil {
					return
				}
				job.call()
				atomic.AddUint64(w.counter, ^uint64(0))
			}
		}
	}()
}

func (w *worker) AddJob(job jobCall) {
	w.task <- job
	return
}