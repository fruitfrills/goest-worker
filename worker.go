package goest_worker

type WorkerInterface interface {
	start()
	getQuitChan() (chan bool)
}

type worker struct {
	task     		chan JobInstance
	quitChan 		chan bool
	workerPool		workerPoolType
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

func (w *worker) start() {
	go func() {
		for {
			select {
			case <-w.quitChan:
				return
			default:
				w.workerPool <- w.task
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
