package redis

import "goest-worker/common"

// local backend
type RedisBackend struct {
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


func (backend *RedisBackend) AddJobToPool(common.PoolInterface, common.JobInstance) () {
	panic(`not implemented`)
	return
}

func (backend *RedisBackend) AddPeriodicJob(common.PoolInterface, common.Job, interface{}, ... interface{}) (job common.PeriodicJob) {
	panic(`not implemented`)
	return
}

func (backend *RedisBackend) Scheduler(common.PoolInterface) {
	panic(`not implemented`)
	return
}

func (backend *RedisBackend) Processor(common.PoolInterface) {
	panic(`not implemented`)
	return
}

func (backend *RedisBackend) Stop(common.PoolInterface) (p common.PoolInterface) {
	panic(`not implemented`)
	return p
}

func (backend *RedisBackend) Start(common.PoolInterface, int) (p common.PoolInterface) {
	panic(`not implemented`)
	return
}
func (backend *RedisBackend) NewJob(common.PoolInterface, interface{}) (job common.Job) {
	panic(`not implemented`)
	return
}