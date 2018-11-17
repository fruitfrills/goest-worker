package goest_worker

import (
	"time"
	"context"
)

type WorkerPoolType chan WorkerInterface

// Factory for creating new queues
type PoolQueue func(ctx context.Context, capacity int) Queue

// Facory for creating new counters
type CounterMiddleware func() Counter

// inner interface
type workerPool interface {
	// getting context
	Context() context.Context

	// run job to pool
	Insert(jobCall) ()

	// create periodic job
	InsertPeriodic(job Job, period interface{}, arguments ... interface{}) (PeriodicJob, error)
}

// inner interface
type jobCall interface {
	// private method for calling jobs
	call()

	// calc priority
	Priority() int
}

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

	// use middlewares, queues and other (TODO: middleware)
	Use(args ... interface{}) Pool
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

type Queue interface {
	// get job
	Pop() (jobCall)

	// Insert job
	Insert(job jobCall)

	// size of queue
	Len() (uint64)
}

type Counter interface {
	// ++i
	Increment()

	// --i
	Decrement()

	// Waiting for counter will be equals to zero or context is closed
	Wait(context.Context)

	// Get counter value
	Len() uint64
}
