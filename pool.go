package goest_worker

import "time"

// channel of channel for balancing tasks between workers
type workerPoolType chan chan TaskInterface

var Pool dispatcher

type dispatcher struct {
	WorkerPool 				chan chan TaskInterface
	TaskQueue  				chan TaskInterface
	WorkersPoolQuitChan 	[]chan bool
}

// create pool of workers
func NewPool(count int) *dispatcher {
	Pool = dispatcher{
		WorkerPool: 	make(workerPoolType, count),
		TaskQueue: 	make(chan TaskInterface),
	}
	for i := 0; i < count; i++ {
		worker := NewWorker(i+1, Pool.WorkerPool)
		Pool.WorkersPoolQuitChan = append(Pool.WorkersPoolQuitChan, worker.QuitChan)
		worker.Start()
	}
	return &Pool
}

// use with go
func (D *dispatcher) AddTask(task TaskInterface){
	D.TaskQueue <- task
}

func (D *dispatcher) addTicker(task TaskInterface, arg interface{})  {
	quitChan := make(chan bool)
	D.WorkersPoolQuitChan = append(D.WorkersPoolQuitChan, quitChan)
	go func() {
		var diff time.Duration;
		switch arg.(type) {
		case time.Time:
			diff = arg.(time.Time).Sub(time.Now())
		case time.Duration:
			diff = arg.(time.Duration)
		case Schedule:
			diff = arg.(*Schedule).Next()
		}
		select {
			case <- time.After(diff):
				Pool.AddTask(task)
			case <-quitChan:
				close(quitChan)
				break
			}
	}()
}

func (D *dispatcher) Start () error {
	go func() {
		for {
			select {
			case task := <-D.TaskQueue:
				worker := <-D.WorkerPool
				worker <- task
			}
		}
	}()
	return nil
}



func (D *dispatcher) Stop() error {
	for _, quit := range D.WorkersPoolQuitChan {
		quit <- true
	}
	close(D.TaskQueue)
	close(D.WorkerPool)
	return nil
}

