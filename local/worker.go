package local

import (
	"goest-worker/common"
)

type worker struct {
	task     		chan common.JobInstance
	quitChan 		chan bool
	workerPool		common.WorkerPoolType
}

func NewWorker(workerPool common.WorkerPoolType) (common.WorkerInterface) {
	return &worker{
		task:     make(chan common.JobInstance),
		quitChan: make(chan bool),
		workerPool: workerPool,
	}
}

func (w *worker) GetQuitChan() (chan bool) {
	return w.quitChan
}


func (w *worker) AddJob(job common.JobInstance) () {
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
