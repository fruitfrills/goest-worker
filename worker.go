package goest_worker

import (
	"sync/atomic"
	"sync"
)


type WorkerInterface interface {
	start()
	getQuitChan() (chan bool)
	addJob(JobInstance) ()
}

type worker struct {
	task     		chan JobInstance
	quitChan 		chan bool
	workerPool		workerPoolType
	state			int32
	sync.Mutex
}

func NewWorker(workerPool workerPoolType) (WorkerInterface) {
	return &worker{
		task:     make(chan JobInstance),
		quitChan: make(chan bool),
		workerPool: workerPool,
	}
}

func (w *worker) getQuitChan() (chan bool) {
	return w.quitChan
}

func (w *worker) setState(state int32) {
	atomic.StoreInt32(&(w.state), int32(state))
}

func (w *worker) addJob(job JobInstance) () {
	w.task <- job
	return
}

func (w *worker) start() {
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
				work.call()
			case <-w.quitChan:
				return
			}
		}
	}()
}
