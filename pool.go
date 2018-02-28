package goest_worker

import (
	"time"
	"sync/atomic"
	"sync"
	"github.com/gorhill/cronexpr"
	"sort"
)

// default pool
var MainPool = New()

// pool states
const (
	POOL_STOPPED int32 = iota
	POOL_STARTED
)

// channel of channel for balancing tasks between workers
type workerPoolType chan WorkerInterface

type PoolInterface interface {
	Start(count int) (PoolInterface)
	Stop() (PoolInterface)
	NewJob(taskFn interface{}) (Job)

	addJobToPool(job JobInstance)
	addPeriodicJob(job Job, period interface{}, arguments ... interface{}) (PeriodicJob)
}

// main wrapper over workers pool
type pool struct {
	PoolInterface

	// can getting free worker from this chan
	workerPool chan WorkerInterface

	// pool's state
	state int32

	// state mutex
	stateMutex *sync.Mutex

	// put job
	jobQueue chan JobInstance

	// slice of quit channel for soft finish
	workersPoolQuitChan []chan bool

	// periodic jobs
	periodicJob []PeriodicJob

	// periodic mutex
	periodicMutex *sync.RWMutex
}

// create pool of workers
func New() (PoolInterface) {
	return &pool{
		stateMutex:    &sync.Mutex{},
		periodicMutex: &sync.RWMutex{},
		state:         POOL_STOPPED,
	}
}

// put job to queue
func (p *pool) addJobToPool(task JobInstance) {
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

func (p *pool) addPeriodicJob(job Job, period interface{}, arguments ... interface{}) (PeriodicJob) {
	p.periodicMutex.Lock()
	defer p.periodicMutex.Unlock()
	var pJob PeriodicJob
	switch period.(type) {
	case time.Duration:
		pJob = &timeDurationPeriodicJob{
			job:      job,
			duration: period.(time.Duration),
			last:     time.Now(),
			args: 	  arguments,
		}
		p.periodicJob = append(p.periodicJob, )
	case string:
		pJob = &cronPeriodicJob{
			job:  job,
			expr: cronexpr.MustParse(period.(string)),
			args: 	  arguments,
		}
	default:
		panic("unknown period")
	}
	p.periodicJob = append(p.periodicJob, pJob)
	return pJob
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

	p.jobQueue = make(chan JobInstance)
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
			case job := <-p.jobQueue:
				// if close channel
				if job == nil {
					return
				}
				var worker WorkerInterface
				// get free worker and send task
				worker = <-p.workerPool
				worker.addJob(job)
			}
		}
	}()

	quitPeriodicChan := make(chan bool)
	p.workersPoolQuitChan = append(p.workersPoolQuitChan, quitPeriodicChan)
	// periodic proccess
	go func() {
		lastCall := time.Now()
		for {
			select {
			case <-quitPeriodicChan:
				return
			default:
				// check all periodic jobs
				if len(p.periodicJob) == 0 {
					<-time.After(time.Second)
					continue
				}
				// copy periodic jobs
				p.periodicMutex.Lock()
				periodicJobs := append([]PeriodicJob(nil), p.periodicJob...)
				p.periodicMutex.Unlock()

				// interval for count tasks
				maxInterval := lastCall.Add(time.Minute)
				queue := []nextJob{}
				JOB_LOOP:
					for _, job := range periodicJobs {
						next := time.Now()
						for {
							next = job.Next(next)
							// drop job if job out of interval
							if next.Sub(maxInterval) > 0 {
								continue JOB_LOOP
							}

							// add job to queue
							queue = append(queue, nextJob{
								job:  job,
								next: next,
							})
						}
					}
				// sort queue by time
				sort.Sort(nextJobSorter(queue))

				for _, job := range queue {
					<-time.After(job.next.Sub(lastCall))
					lastCall = time.Now()
					job.job.run()
				}
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
func (p *pool) NewJob(taskFn interface{}) (job Job) {
	job = NewJob(taskFn)
	job.(*jobFunc).pool = p
	return
}
