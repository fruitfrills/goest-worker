package local

import (
	goestworker "goest-worker"
)

type worker struct {
	task     		chan goestworker.JobInstance
	quitChan 		chan bool
	workerPool		workerPoolType
}

func NewWorker(workerPool workerPoolType) (goestworker.WorkerInterface) {
	return &worker{
		task:     make(chan goestworker.JobInstance),
		quitChan: make(chan bool),
		workerPool: workerPool,
	}
}

func (w *worker) GetQuitChan() (chan bool) {
	return w.quitChan
}


func (w *worker) AddJob(job goestworker.JobInstance) () {
	w.task <- job
	return
}

func (w *worker) Start() {
	go func() {
		for {
			select {
			case <-w.quitChan:
				return
			default:
				w.workerPool <- w
			}
			select {
			case work := <-w.task:
				if work == nil {
					return
				}
				work.Call()
			case <-w.quitChan:
				return
			}
		}
	}()
}
