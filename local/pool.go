package local

import (
	"time"
	"github.com/gorhill/cronexpr"
	"sort"
	goestworker "goest-worker/common"
	"reflect"
)


// channel of channel for balancing tasks between workers
type workerPoolType chan goestworker.WorkerInterface

// local backend
type LocalBackend struct {
	goestworker.PoolBackendInterface

	// can getting free worker from this chan
	workerPool chan goestworker.WorkerInterface

	// put job
	jobQueue chan goestworker.JobInstance

	// slice of quit channel for soft finish
	workersPoolQuitChan []chan bool

	// periodic jobs
	periodicJob []goestworker.PeriodicJob

}

// put job to queue
func (backend *LocalBackend) AddJobToPool(p goestworker.PoolInterface, task goestworker.JobInstance) () {
	p.Lock()
	defer p.Unlock()

	// check state
	// is pool is stopped, job is dropped ...
	if p.IsStopped() {
		task.Drop()
		return
	}
	backend.jobQueue <- task
}

func (backend *LocalBackend) AddPeriodicJob(p goestworker.PoolInterface, job goestworker.Job, period interface{}, arguments ... interface{}) (goestworker.PeriodicJob) {
	var pJob goestworker.PeriodicJob
	switch period.(type) {
	case time.Duration:
		pJob = &timeDurationPeriodicJob{
			job:      job,
			duration: period.(time.Duration),
			last:     time.Now(),
			args: 	  arguments,
		}
		backend.periodicJob = append(backend.periodicJob, pJob)
	case string:
		pJob = &cronPeriodicJob{
			job:  job,
			expr: cronexpr.MustParse(period.(string)),
			args: 	  arguments,
		}
	default:
		panic("unknown period")
	}
	backend.periodicJob = append(backend.periodicJob, pJob)
	return pJob
}

func (backend *LocalBackend) Processor(goestworker.PoolInterface) {
	for {
		select {
		case job := <-backend.jobQueue:
			// if close channel
			if job == nil {
				return
			}
			var worker goestworker.WorkerInterface
			// get free worker and send task
			worker = <-backend.workerPool
			worker.AddJob(job)
		}
	}
}

func (backend *LocalBackend) Scheduler (p goestworker.PoolInterface) {
	quitPeriodicChan := make(chan bool)
	backend.workersPoolQuitChan = append(backend.workersPoolQuitChan, quitPeriodicChan)
	lastCall := time.Now()
	periodicJobs := append([]goestworker.PeriodicJob(nil), backend.periodicJob...)
	for {
		select {
		case <-quitPeriodicChan:
			return
		default:
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

func (backend *LocalBackend) Start(p goestworker.PoolInterface, count int) (goestworker.PoolInterface) {

	// if dispatcher started - do nothing
	if !p.IsStopped() {
		return p
	}

	backend.jobQueue = make(chan goestworker.JobInstance)
	backend.workerPool = make(workerPoolType, count)
	for i := 0; i < count; i++ {
		worker := NewWorker(backend.workerPool)
		backend.workersPoolQuitChan = append(backend.workersPoolQuitChan, worker.GetQuitChan())
		worker.Start()
	}
	// main process
	go backend.Processor(p)


	// periodic proccess

	if  len(backend.periodicJob) != 0 {
		go backend.Scheduler(p)
	}

	// set state to started
	return p
}

func (backend *LocalBackend) Stop(p goestworker.PoolInterface) (goestworker.PoolInterface) {

	// if dispatcher stopped - do nothing
	if p.IsStopped() {
		return p
	}

	// send close to all quit channels
	for _, quit := range backend.workersPoolQuitChan {
		close(quit)
	}

	backend.workersPoolQuitChan = [](chan bool){};
	close(backend.jobQueue)
	close(backend.workerPool)
	return p
}

// create simple jobs
func (backend *LocalBackend) NewJob(p goestworker.PoolInterface, taskFn interface{}) (goestworker.Job) {

	fn := reflect.ValueOf(taskFn)
	fnType := fn.Type()

	if fnType.Kind() != reflect.Func {
		panic("job is not func")
	}

	return &jobFunc{
		fn: fn,
		maxRetry: -1,
		pool: p,
	}
}