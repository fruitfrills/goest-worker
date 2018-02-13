package goest_worker

import (
	"time"
)

// channel of channel for balancing tasks between workers
type workerPoolType chan chan Job

var Pool dispatcher

type DispatcherInterface interface {
	Start() (DispatcherInterface)
	Stop() (DispatcherInterface)
	AddTask(job Job)
	addTicker(job Job, arg interface{})
}

// main wrapper over workers pool
type dispatcher struct {
	DispatcherInterface

	// can getting free worker from this chann
	workerPool          chan chan Job

	// put job
	jobQueue           chan Job

	// slice of quit channel for soft finish
	workersPoolQuitChan []chan bool
}

// create pool of workers
func NewPool(count int) (DispatcherInterface) {
	Pool = dispatcher{
		workerPool: make(workerPoolType, count),
		jobQueue:  make(chan Job),
	}

	// use numCPU
	for i := 0; i < count; i++ {
		worker := NewWorker()
		Pool.workersPoolQuitChan = append(Pool.workersPoolQuitChan, worker.getQuitChan())
		worker.start()
	}
	return &Pool
}

// put task to queue
func (D *dispatcher) addTask(task Job) {
	D.jobQueue <- task
}

// create periodic tasks
func (D *dispatcher) addTicker(task Job, arg interface{}) {
	quitChan := make(chan bool)
	D.workersPoolQuitChan = append(D.workersPoolQuitChan, quitChan)
	go func() {
		for {
			var once bool; // run once at time
			var diff time.Duration;
			switch arg.(type) {
			case time.Time:
				diff = arg.(time.Time).Sub(time.Now())
				once = true // run at time
			case time.Duration:
				diff = arg.(time.Duration)
			case *Schedule:
				diff = arg.(*Schedule).Next()
			}
			select {
			case <- quitChan:
				return 
			case <-time.After(diff):
				task.Run().Wait()
			}
			// break
			if once{
				return
			}
		}
	}()
}

func (D *dispatcher) Start() (DispatcherInterface) {
	go func() {
		for {
			select {
			case task := <-D.jobQueue:
				// if close channel
				if task == nil { 
					return
				}
				// get worker and send task
				worker := <-D.workerPool
				worker <- task
			}
		}
	}()
	return D
}

// close all channels, stopping workers and periodic tasks
func (D *dispatcher) Stop() (DispatcherInterface) {
	for _, quit := range D.workersPoolQuitChan {
		close(quit)
	}
	D.workersPoolQuitChan = [](chan bool){};
	close(D.jobQueue)
	close(D.workerPool)
	return D
}
