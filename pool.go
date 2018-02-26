package goest_worker

import (
	"time"
	"sync/atomic"
	"sync"
    "github.com/gorhill/cronexpr"
)

// default pool
var MainPool = New()

// pool states
const (
	POOL_STOPPED int32 = iota
	POOL_STARTED
)

// channel of channel for balancing tasks between workers
type workerPoolType chan chan Job

type PoolInterface interface {
	Start(count int) (PoolInterface)
	Stop() (PoolInterface)
	NewJob(taskFn interface{}, arguments ... interface{}) (Job, error)

	AddTask(job Job)
	addTicker(job Job, arg interface{})
}

// main wrapper over workers pool
type pool struct {
	PoolInterface

	// can getting free worker from this chan
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
func New() (PoolInterface) {
	return &pool{
		stateMutex: &sync.Mutex{},
		state: POOL_STOPPED,
	}
}

// put job to queue
func (p *pool) AddTask(task Job) {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	// check state
	// is pool is stopped, job is dropped ...
	if p.isStopped() {
		task.drop()
		return
	}
	p.jobQueue <- task
}

// create periodic tasks
// TODO: use list of periodic tasks
func (p *pool) addTicker(task Job, arg interface{}) {
	quitChan := make(chan bool)
	p.workersPoolQuitChan = append(p.workersPoolQuitChan, quitChan)
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
func (p *pool) isStopped() (bool) {
	return atomic.LoadInt32(&(p.state)) == POOL_STOPPED
}

// set state
func (p *pool) setState(state int32) () {
	atomic.StoreInt32(&(p.state), int32(state))
}

// start dispatcher
func (p *pool) Start(count int) (PoolInterface) {
	// lock for start
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	// if dispatcher started - do nothing
	if !p.isStopped() {
		return p
	}

	p.jobQueue = make(chan Job)
	p.workerPool = make(workerPoolType, count)
	for i := 0; i < count; i++ {
		worker := NewWorker(p.workerPool)
		p.workersPoolQuitChan = append(p.workersPoolQuitChan, worker.getQuitChan())
		worker.start()
	}
	// main process
	go func() {
		for {
			select {
			case task := <-p.jobQueue:
				// if close channel
				if task == nil {
					return
				}
				// get worker and send task
				worker := <-p.workerPool
				worker <- task
			}
		}
	}()

	// set state to started
	p.setState(POOL_STARTED)
	return p
}

// close all channels, stopping workers and periodic tasks
// if the dispatcher is restarted, all tasks will be stoppep. You need starts tasks again ...
func (p *pool) Stop() (PoolInterface) {

	// lock for stop
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()
	// if dispatcher stopped - do nothing
	if p.isStopped() {
		return p
	}
	// send close to all quit channels
	for _, quit := range p.workersPoolQuitChan {
		close(quit)
	}
	p.workersPoolQuitChan = [](chan bool){};
	close(p.jobQueue)
	close(p.workerPool)

	// set state to stopped
	p.setState(POOL_STOPPED)
	return p
}

// create job for current pool
func (p *pool) NewJob(taskFn interface{}, arguments ... interface{})(task Job, err error){
	task, err = NewJob(taskFn, arguments ... )
	if err != nil {
		return

	}
	task.(*jobFunc).pool = p
	return
}