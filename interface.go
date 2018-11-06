package goest_worker

import (
	"time"
	"context"
)

type workerPool interface {
	Context() context.Context
	AddJobToPool(jobCall) ()
	AddPeriodicJob(job Job, period interface{}, arguments ... interface{}) (PeriodicJob, error)
	Done()
}

type jobCall interface {
	Call()
}

type WorkerPoolType chan WorkerInterface

type GoestWorker interface {
	Start(ctx context.Context, count int) GoestWorker
	Stop()
	Context() context.Context
	NewJob(taskFn interface{}) (Job)
	Wait()
}

type Job interface {
	Bind(val bool) Job
	Run(args ... interface{}) JobInstance
	RunEvery(period interface{}, args ... interface{}) (PeriodicJob, error)
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