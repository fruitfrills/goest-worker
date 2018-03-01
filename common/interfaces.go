package common

import (
	"time"
	"sync"
)

type PoolBackendInterface interface {

	// add jobbs
	AddJobToPool(PoolInterface, JobInstance) ()

	// add periodic jobs
	AddPeriodicJob(PoolInterface, Job, interface{}, ... interface{}) (PeriodicJob)

	// start process for calculate tasks
	Scheduler(PoolInterface)

	//process pool
	Processor(PoolInterface)

	// stop pool
	Stop(PoolInterface) PoolInterface

	// start pool
	Start(PoolInterface, int) PoolInterface

	// create job on backend
	NewJob(PoolInterface, interface{}) (Job)
}

type PoolInterface interface {
	sync.Locker
	// start worker pool, count - count of workers
	Start(count int) (PoolInterface)

	// stop all worker
	Stop() (PoolInterface)

	// create new job for current pool
	NewJob(taskFn interface{}) (Job)

	// put job to queue
	AddJobToPool(job JobInstance)

	// put job to scheduler
	AddPeriodicJob(job Job, period interface{}, arguments ... interface{}) (PeriodicJob)

	// check stop
	IsStopped() bool

}

type Job interface {

	// run job
	Run(args ... interface{}) JobInstance

	// create periodic job
	RunEvery(period interface{}, args ... interface{}) PeriodicJob

	// use job instance as first argument
	Bind(bool) Job

	// count max retry
	SetMaxRetry(int) Job

	// set pool for work
	SetPool(PoolInterface) Job
}

type JobInstance interface {

	// waitig results
	Wait() JobInstance

	// get results
	Result() ([]interface{}, error)

	// call
	Call() JobInstance

	// drop
	Drop() JobInstance
}

type JobInjection interface {

	// repeat job
	Retry() JobInstance
}

type PeriodicJob interface {

	// calculate next time of jpb
	Next(time.Time) time.Time

	// run periodic job
	Run() ()
}

type WorkerInterface interface {

	// start proccess
	Start()

	// get quit channel for soft exit
	GetQuitChan() (chan bool)

	// add job to work
	AddJob(JobInstance) ()
}
