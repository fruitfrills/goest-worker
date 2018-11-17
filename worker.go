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
}

func newWorker(workerPool WorkerPoolType, counter Counter) (WorkerInterface)  {
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
				if w.counter != nil {
					w.counter.Decrement()
				}
			}
		}
	}()
}

func (w *worker) AddJob(job jobCall) {
	w.task <- job
	return
}