package local

import (
	"time"
	"sync/atomic"
	"sync"
	"github.com/gorhill/cronexpr"
	"sort"
	goestworker "goest-worker"
)

// channel of channel for balancing tasks between workers
type workerPoolType chan goestworker.WorkerInterface

// default pool
var MainPool = New()

// main wrapper over workers pool
type pool struct {
	goestworker.PoolInterface

	// can getting free worker from this chan
	workerPool chan goestworker.WorkerInterface

	// pool's state
	state int32

	// state mutex
	stateMutex *sync.Mutex

	// put job
	jobQueue chan goestworker.JobInstance

	// slice of quit channel for soft finish
	workersPoolQuitChan []chan bool

	// periodic jobs
	periodicJob []goestworker.PeriodicJob
}

// create pool of workers
func New() (goestworker.PoolInterface) {
	return &pool{
		stateMutex:    &sync.Mutex{},
		state:         goestworker.POOL_STOPPED,
	}
}

// put job to queue
func (p *pool) AddJobToPool(task goestworker.JobInstance) {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	// check state
	// is pool is stopped, job is dropped ...
	if p.isStopped() {
		task.Drop()
		return
	}
	p.jobQueue <- task
}

func (p *pool) AddPeriodicJob(job goestworker.Job, period interface{}, arguments ... interface{}) (goestworker.PeriodicJob) {
	var pJob goestworker.PeriodicJob
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
	return atomic.LoadInt32(&(p.state)) == goestworker.POOL_STOPPED
}

// set state
func (p *pool) setState(state int32) () {
	atomic.StoreInt32(&(p.state), int32(state))
}

// start dispatcher
func (p *pool) Start(count int) (goestworker.PoolInterface) {
	// lock for start
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	// if dispatcher started - do nothing
	if !p.isStopped() {
		return p
	}

	p.jobQueue = make(chan goestworker.JobInstance)
	p.workerPool = make(workerPoolType, count)
	for i := 0; i < count; i++ {
		worker := NewWorker(p.workerPool)
		p.workersPoolQuitChan = append(p.workersPoolQuitChan, worker.GetQuitChan())
		worker.Start()
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
				var worker goestworker.WorkerInterface
				// get free worker and send task
				worker = <-p.workerPool
				worker.AddJob(job)
			}
		}
	}()


	// periodic proccess

	scheduler := func() {
		quitPeriodicChan := make(chan bool)
		p.workersPoolQuitChan = append(p.workersPoolQuitChan, quitPeriodicChan)
		lastCall := time.Now()
		periodicJobs := append([]goestworker.PeriodicJob(nil), p.periodicJob...)
		for {
			select {
			case <-quitPeriodicChan:
				return
			default:
				//// check all periodic jobs
				//if len(p.periodicJob) == 0 {
				//	<-time.After(time.Second)
				//	continue
				//}
				//// copy periodic jobs
				//p.periodicMutex.Lock()
				//periodicJobs := append([]goestworker.PeriodicJob(nil), p.periodicJob...)
				//p.periodicMutex.Unlock()

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

				for _, pJob := range queue {
					select {
					case <-time.After(pJob.next.Sub(lastCall)):
						lastCall = time.Now()
						pJob.job.Run()
					case <- quitPeriodicChan:
						return
					}

				}
			}

		}
	}

	if  len(p.periodicJob) != 0 {
		go scheduler()
	}

	// set state to started
	p.setState(goestworker.POOL_STARTED)
	return p
}

// close all channels, stopping workers and periodic tasks
// if the dispatcher is restarted, all tasks will be stoppep. You need starts tasks again ...
func (p *pool) Stop() (goestworker.PoolInterface) {

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
	p.setState(goestworker.POOL_STOPPED)
	return p
}

// create job for current pool
func (p *pool) NewJob(taskFn interface{}) (job goestworker.Job) {
	job = NewJob(taskFn)
	job.(*jobFunc).pool = p
	return
}
