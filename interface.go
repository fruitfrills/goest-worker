package goest_worker

import (
	"time"
	"context"
)

type workerPool interface {

	// getting context
	Context() context.Context

	// run job to pool
	AddJobToPool(jobCall) ()

	// create periodic job
	AddPeriodicJob(job Job, period interface{}, arguments ... interface{}) (PeriodicJob, error)
}

type jobCall interface {

	// private method for calling jobs
	call()

	// calc priority
	Priority() int
}

type WorkerPoolType chan WorkerInterface

type Pool interface {

	// Start pool
	Start(ctx context.Context, count int) Pool

	// Stop pool
	Stop()

	// getting context
	Context() context.Context

	// Create new job from any function
	NewJob(taskFn interface{}) (Job)

	// Waiting for run all functions or <- context.Done()
	Wait()
}

type Job interface {

	// Set the first argument to current JobInstance
	Bind(val bool) Job

	// Put current job to pool
	Run(args ... interface{}) JobInstance

	// Put current job to pool every
	// time.Duration
	// cron expression
	RunEvery(period interface{}, args ... interface{}) (PeriodicJob, error)

	// Put current job to pool with priority
	RunWithPriority(priority int, args ... interface{}) JobInstance
}

type JobInstance interface {

	// Get context
	Context() context.Context

	// Cancel job
	Cancel()

	// Get results
	Result() ([]interface{}, error)

	// Wait (<-context.Done())
	Wait() JobInstance
}


type PeriodicJob interface {

	// get next time to running
	Next(time.Time) time.Time

	// Put current job to pool
	Run() ()
}

type WorkerInterface interface {

	// Start worker
	Start(ctx context.Context)

	// Put job to worker
	AddJob(jobCall) ()
}