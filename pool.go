package goest_worker


import (
	"time"
)
// channel of channel for balancing tasks between workers
type workerPoolType chan chan *Task

var Pool dispatcher

type dispatcher struct {
	WorkerPool 				chan chan *Task
	TaskQueue  				chan *Task
	WorkersPoolQuitChan 	[]chan bool
	PeriodicTasks 			[]PeriodicTask
}

// create pool of workers
func NewPool(count int) *dispatcher {
	Pool = dispatcher{
		WorkerPool: 	make(workerPoolType),
		TaskQueue: 	make(chan *Task),
	}
	for i := 0; i < count; i++ {
		worker := NewWorker(i+1, Pool.WorkerPool)
		Pool.WorkersPoolQuitChan = append(Pool.WorkersPoolQuitChan, worker.QuitChan)
		worker.Start()
	}
	return &Pool
}

// use with go
func (D *dispatcher) AddTask(task *Task){
	D.TaskQueue <- task
}

// create periodic task and append its to dispatcher
func (D *dispatcher) AddPeriodicTask(interval time.Duration, task Task){
	D.PeriodicTasks = append(D.PeriodicTasks, PeriodicTask{
		Task: task,
		Period: interval,
	})
}


func (D *dispatcher) Run() {
	for {
		select {
		case task := <-D.TaskQueue:
			worker := <-D.WorkerPool
			worker <- task
		}
	}
}


func (D *dispatcher) Start () error {
	go D.Run()
	go D.periodicTasksProcessor()
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



func (D *dispatcher) periodicTasksProcessor(){
	for _, task := range D.PeriodicTasks{
		quitChan := make(chan bool)

		// add to begin
		D.WorkersPoolQuitChan = append([](chan bool){quitChan,}, D.WorkersPoolQuitChan ... )
		go func(t PeriodicTask, q chan bool) {
			for {
				select {
				case Pool.TaskQueue <- &t.Task:
				case <- q:
					return
				}
				select {
				case <-time.After(t.Period):
					continue
				case <- q:
					return
				}
			}
		}(task, quitChan)
	}
}
