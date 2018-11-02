package goest_worker

import (
	"context"
	"sync"
)

type worker struct {
	task 				chan jobCall
	pool				WorkerPoolType
	wg 					*sync.WaitGroup
}

func newWorker(workerPool WorkerPoolType, wg *sync.WaitGroup) (WorkerInterface)  {
	return &worker{
		task: make(chan jobCall),
		pool: workerPool,
		wg: wg,
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
				job.Call()
				w.wg.Done()
			}
		}
	}()
}

func (w *worker) AddJob(job jobCall) {
	w.task <- job
	return
}