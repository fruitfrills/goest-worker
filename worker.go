package goest_worker

type worker struct {
	ID         	int
	WorkerPool 	workerPoolType
	Task       	chan *Task
	QuitChan   	chan bool
}

func NewWorker(id int, workerQueue chan chan *Task) *worker {
	return &worker{
		ID:         id,
		WorkerPool: workerQueue,
		Task:       make(chan *Task),
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
			work.Call()
		case <- w.QuitChan:
			return
		}
	}
}
