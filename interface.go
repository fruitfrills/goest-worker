package goest_worker

import (
	"time"
	"context"
)

// Channel for free workers
type workerPoolType chan iWorker

// Factory for creating new queues
type PoolQueue func(ctx context.Context, capacity int) Queue

// Factory for creating new counters
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

// Pool defines pool interface required for use on client-side
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

	// change the numbers of workers
	Resize(count int) Pool
}

// Job defines job interface required for use on client side
type Job interface {
	// Set the first argument to current JobInstance
	Bind(val bool) Job

	// Put current job to pool
	Run(args ... interface{}) JobInstance

	// Put current job to pool every, use:
	//
	//    time.Duration
	//
	// or
	//
	//    "0 */5 * ? * *" // cron expression
	//
	RunEvery(period interface{}, args ... interface{}) (PeriodicJob, error)

	// Put current job to pool with priority
	RunWithPriority(priority int, args ... interface{}) JobInstance
}

// JobInstance defines current job instance interface required for use on client side
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

//
type PeriodicJob interface {
	// get next time to running
	Next(time.Time) time.Time

	// Put current job to pool
	Run() ()
}

// iWorker defines all worker interface required for use in the pool
type iWorker interface {
	// Start worker
	Start()

	// Put job to worker
	AddJob(jobCall) ()

	// Get context
	Context() (ctx context.Context)

	// Close worker
	Cancel()
}

// Queue defines all queue interface required for use in the pool
type Queue interface {
	// get job
	Pop() (jobCall)

	// Insert job
	Insert(job jobCall)

	// size of queue
	Len() (uint64)

	// change capacity
	SetCapacity(ctx context.Context, capacity int)
}

// The counter is used to count the tasks in the pool.
// Used to wait for all tasks to complete.
type Counter interface {
	// ++i
	Increment()

	// --i
	Decrement()

	// Waits while counter value will be equals to zero or context will be done
	Wait(context.Context)

	// Get counter value
	Len() uint64
}
