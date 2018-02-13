package goest_worker

type WorkerInterface interface {
	start()
	getQuitChan() (chan bool)
}

type worker struct {
	task     chan Job
	quitChan chan bool
}

func NewWorker() (WorkerInterface) {
	return &worker{
		task:     make(chan Job),
		quitChan: make(chan bool),
	}
}

func (w *worker) getQuitChan() (chan bool) {
	return w.quitChan
}

func (w *worker) start() {
	go func() {
		for {
			select {
			case <-w.quitChan:
				return
			default:
				Pool.workerPool <- w.task
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
