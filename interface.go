package goest_worker

import (
	"time"
	"context"
)

// channel of channel for balancing tasks between workers
type WorkerPoolType chan WorkerInterface

type workerPool interface {
	Context() context.Context
	AddJobToPool(jobCall) ()
	Done()
	AddPeriodicJob(job Job, period interface{}, arguments ... interface{}) (PeriodicJob, error)
}


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

// private interaface for job
type jobCall interface {
	Call()
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