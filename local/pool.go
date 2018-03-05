package local

import (
	"time"
	"sort"
	"goest-worker/common"
	"reflect"
)

// local backend
type LocalBackend struct {
	common.PoolBackendInterface

	// can getting free worker from this chan
	workerPool chan common.WorkerInterface

	// put job
	jobQueue chan common.JobInstance

	// slice of quit channel for soft finish
	workersPoolQuitChan []chan bool

	// periodic jobs
	periodicJob []common.PeriodicJob

}

// put job to queue
func (backend *LocalBackend) AddJobToPool(p common.PoolInterface, task common.JobInstance) () {
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

func (backend *LocalBackend) AddPeriodicJob(p common.PoolInterface, job common.Job, period interface{}, arguments ... interface{}) (common.PeriodicJob) {
	p.Lock()
	defer p.Unlock()

	if !p.IsStopped() {
		panic(common.ErrorJobAdding)
	}

	var pJob common.PeriodicJob
	switch period.(type) {
	case time.Duration:
		pJob = common.NewTimeDurationJob(job, period.(time.Duration), arguments ... )
	case string:
		pJob = common.NewCronJob(job, period.(string), arguments ...)
	default:
		panic("unknown period")
	}


	backend.periodicJob = append(backend.periodicJob, pJob)
	return pJob
}

func (backend *LocalBackend) Processor(common.PoolInterface) {
	quitChan := make(chan bool)
	backend.workersPoolQuitChan = append(backend.workersPoolQuitChan, quitChan)
	for {
		select {
		case job := <-backend.jobQueue:
			// if close channel
			if job == nil {
				return
			}
			var worker common.WorkerInterface
			// get free worker and send task
			worker = <-backend.workerPool
			worker.AddJob(job)
		case <- quitChan:
			return
		}
	}
}

func (backend *LocalBackend) Scheduler (p common.PoolInterface) {
	quitPeriodicChan := make(chan bool)
	backend.workersPoolQuitChan = append(backend.workersPoolQuitChan, quitPeriodicChan)
	lastCall := time.Now()
	periodicJobs := append([]common.PeriodicJob(nil), backend.periodicJob...)
	for {
		select {
		case <-quitPeriodicChan:
			return
		default:
			maxInterval := lastCall.Add(time.Minute)
			queue := []common.NextJob{}
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
						queue = append(queue, common.NextJob{
							Job:  job,
							Next: next,
						})
					}
				}
			// sort queue by time
			sort.Sort(common.NextJobSorter(queue))

			for _, pJob := range queue {
				select {
				case <-time.After(pJob.Next.Sub(lastCall)):
					lastCall = time.Now()
					pJob.Job.Run()
				case <- quitPeriodicChan:
					return
				}

			}
		}

	}
}

func (backend *LocalBackend) Start(p common.PoolInterface, count int) (common.PoolInterface) {

	// if dispatcher started - do nothing
	if !p.IsStopped() {
		return p
	}

	backend.jobQueue = make(chan common.JobInstance)
	backend.workerPool = make(common.WorkerPoolType, count)
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

func (backend *LocalBackend) Stop(p common.PoolInterface) (common.PoolInterface) {

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
func (backend *LocalBackend) NewJob(p common.PoolInterface, taskFn interface{}) (common.Job) {

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

func (backend *LocalBackend) Register (p common.PoolInterface, name string, taskFn interface{}) (common.Job) {
	return backend.NewJob(p, taskFn)
}