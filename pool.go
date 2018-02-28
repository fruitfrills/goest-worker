package goest_worker


type PoolInterface interface {

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
}

