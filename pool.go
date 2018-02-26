package goest_worker

import (
	"time"
	"sync/atomic"
	"sync"
    "github.com/gorhill/cronexpr"
)


// pool states
const (
	POOL_STOPPED int32 = iota
	POOL_STARTED
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
	workerPool chan chan Job

	// pool's state
	state int32

	// state mutex
	stateMutex *sync.Mutex

	// put job
	jobQueue chan Job

	// slice of quit channel for soft finish
	workersPoolQuitChan []chan bool
}

// create pool of workers
func NewPool(count int) (DispatcherInterface) {
	Pool = dispatcher{
		workerPool: make(workerPoolType, count),
		jobQueue:   make(chan Job),
		stateMutex: &sync.Mutex{},
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

	D.stateMutex.Lock()
	defer D.stateMutex.Unlock()

	// check state
	// is disptatcher is stopped, job is dropped ...
	if D.isStopped() {
		task.drop()
		return
	}
	D.jobQueue <- task
}

// create periodic tasks
// TODO: use list of periodic tasks
func (D *dispatcher) addTicker(task Job, arg interface{}) {
	quitChan := make(chan bool)
	D.workersPoolQuitChan = append(D.workersPoolQuitChan, quitChan)
	go func() {
		for {

			var (
				once bool; // run once at time
				diff time.Duration
				now = time.Now()
			)

			switch arg.(type) {
			case time.Time:
				diff = arg.(time.Time).Sub(now)
				once = true // run at time
			case time.Duration:
				diff = arg.(time.Duration)
			case string:
				diff = cronexpr.MustParse(arg.(string)).Next(now).Sub(now)
			}
			select {
			case <-quitChan:
				return
			case <-time.After(diff):
				task.Run()
			}
			// break
			if once {
				return
			}
		}
	}()
}

// check state
func (D *dispatcher) isStopped() (bool) {
	return atomic.LoadInt32(&(D.state)) == POOL_STOPPED
}

// set state
func (D *dispatcher) setState(state int32) () {
	atomic.StoreInt32(&(D.state), int32(state))
}

// start dispatcher
func (D *dispatcher) Start() (DispatcherInterface) {

	// lock for start
	D.stateMutex.Lock()
	defer D.stateMutex.Unlock()

	// if dispatcher started - do nothing
	if !D.isStopped() {
		return D
	}

	// main process
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

	// set state to started
	D.setState(POOL_STARTED)
	return D
}

// close all channels, stopping workers and periodic tasks
// if the dispatcher is restarted, all tasks will be stopped. You need starts tasks again ...
func (D *dispatcher) Stop() (DispatcherInterface) {

	// lock for stop
	D.stateMutex.Lock()
	defer D.stateMutex.Unlock()
	// if dispatcher stopped - do nothing
	if D.isStopped() {
		return D
	}
	// send close to all quit channels
	for _, quit := range D.workersPoolQuitChan {
		close(quit)
	}
	D.workersPoolQuitChan = [](chan bool){};
	close(D.jobQueue)
	close(D.workerPool)

	// set state to stopped
	D.setState(POOL_STOPPED)
	return D
}
