package goest_worker

import (
	"time"
	"fmt"
)

// channel of channel for balancing tasks between workers
type workerPoolType chan chan TaskInterface

var Pool dispatcher

type dispatcher struct {
	WorkerPool          chan chan TaskInterface
	TaskQueue           chan TaskInterface
	WorkersPoolQuitChan []chan bool
}

// create pool of workers
func NewPool(count int) *dispatcher {
	Pool = dispatcher{
		WorkerPool: make(workerPoolType, count),
		TaskQueue:  make(chan TaskInterface),
	}
	for i := 0; i < count; i++ {
		worker := NewWorker()
		Pool.WorkersPoolQuitChan = append(Pool.WorkersPoolQuitChan, worker.getQuitChan())
		worker.start()
	}
	return &Pool
}

// use with go
func (D *dispatcher) AddTask(task TaskInterface) {
	D.TaskQueue <- task
}

func (D *dispatcher) addTicker(task TaskInterface, arg interface{}) {
	quitChan := make(chan bool)
	D.WorkersPoolQuitChan = append(D.WorkersPoolQuitChan, quitChan)
	go func() {
		for {
			var once bool;
			var diff time.Duration;
			switch arg.(type) {
			case time.Time:
				diff = arg.(time.Time).Sub(time.Now())
				once = true // run at time
			case time.Duration:
				diff = arg.(time.Duration)
			case *Schedule:
				diff = arg.(*Schedule).Next()
				fmt.Println(diff)
			}
			select {
			case <- quitChan:
				return 
			case <-time.After(diff):
				task.Run().Wait()
			}
			if once{
				return
			}
		}
	}()
}

func (D *dispatcher) Start() (*dispatcher) {
	go func() {
		for {
			select {
			case task := <-D.TaskQueue:
				if (task == nil) {
					return
				}
				worker := <-D.WorkerPool
				worker <- task
			}
		}
	}()
	return D
}

func (D *dispatcher) Stop() {
	for _, quit := range D.WorkersPoolQuitChan {
		close(quit)
	}
	D.WorkersPoolQuitChan = [](chan bool){};
	close(D.TaskQueue)
	close(D.WorkerPool)
	return
}
