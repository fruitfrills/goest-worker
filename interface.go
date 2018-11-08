package goest_worker

import (
	"time"
	"context"
)

type workerPool interface {
	Context() context.Context
	AddJobToPool(jobCall) ()
	AddPeriodicJob(job Job, period interface{}, arguments ... interface{}) (PeriodicJob, error)
}

type jobCall interface {
	call()
	Priority() int
}

type WorkerPoolType chan WorkerInterface

type Pool interface {
	Start(ctx context.Context, count int) Pool
	Stop()
	Context() context.Context
	NewJob(taskFn interface{}) (Job)
	Wait()
}

type Job interface {
	Bind(val bool) Job
	Run(args ... interface{}) JobInstance
	RunEvery(period interface{}, args ... interface{}) (PeriodicJob, error)
	RunWithPriority(priority int, args ... interface{}) JobInstance
}

type JobInstance interface {
	Context() context.Context
	Cancel()
	Result() ([]interface{}, error)
	Wait() JobInstance
}


type PeriodicJob interface {
	Next(time.Time) time.Time
	Run() ()
}

type WorkerInterface interface {
	Start(ctx context.Context)
	AddJob(jobCall) ()
}