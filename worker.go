package goest_worker

type worker struct {
	ID         	int
	WorkerPool 	workerPoolType
	Task       	chan TaskInterface
	QuitChan   	chan bool
}

func NewWorker(id int, workerQueue chan chan TaskInterface) *worker {
	return &worker{
		ID:         id,
		WorkerPool: workerQueue,
		Task:       make(chan TaskInterface),
		QuitChan:   make(chan bool),
	}
}

func (w *worker) Start() {
	go w.tasksProcessor()
}

func (w *worker) tasksProcessor() {
	for {
		w.WorkerPool <- w.Task
		select {
		case work := <-w.Task:
			work.call()
		case <- w.QuitChan:
			return
		}
	}
}
