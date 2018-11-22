package goest_worker

import (
	"context"
	"time"
	"sort"
	"reflect"
	"errors"
)

// Pool send to queue job from client
type pool struct {

	// inner context for worker's pool
	ctx context.Context

	// Cancel context func
	cancel context.CancelFunc

	// Chan for free workers
	workerPool chan WorkerInterface

	// Periodic jobs
	periodicJob []PeriodicJob

	// factory for create new queue
	prepareQueue PoolQueue

	// any implematation of queue
	queue Queue

	// factory for create new counter
	prepareCounter CounterMiddleware

	// active job counter
	counter Counter

	// active workers
	workers []WorkerInterface

}

// Run worker pool
func (backend *pool) Start(ctx context.Context, count int) Pool {


	// create new context
	backend.ctx, backend.cancel = context.WithCancel(ctx)

	// set default queue
	if backend.prepareQueue == nil {
		backend.prepareQueue = PriorityQueue
	}

	// set counter if needed
	if backend.prepareCounter != nil {
		backend.counter = backend.prepareCounter()
	}

	// create empty queue
	backend.queue = backend.prepareQueue(ctx, count)

	// create worker pool
	backend.workerPool = make(WorkerPoolType, count)
	// create workers
	for i := 0; i < count; i++ {
		worker := newWorker(ctx, backend.workerPool, backend.counter)
		backend.workers = append(backend.workers, worker)
		worker.Start()
	}

	// run processor
	processor(ctx, backend.queue, backend.workerPool)

	// run periodic processor
	if len(backend.periodicJob) != 0 {
		periodicProcessor(ctx, backend.periodicJob)
	}

	return backend
}

func (backend *pool) Resize(count int) Pool {
	if len(backend.workers) == count {
		return backend
	}
	// Reducing the number of workers
	if len(backend.workers) > count {
		for i := len(backend.workers); i > count; i-- {
			var worker WorkerInterface
			worker, backend.workers = backend.workers[0], backend.workers[1:]
			worker.Cancel()
		}

		return backend
	} else {
		// Increasing the number of workers
		for i := len(backend.workers);  i < count; i++ {
			worker := newWorker(backend.ctx, backend.workerPool, backend.counter)
			worker.Start()
			backend.workers = append(backend.workers, worker)
		}
	}
	backend.queue.SetCapacity(backend.ctx, count)
	return backend
}

func (backend *pool) Wait() {
	if backend.counter ==  nil {
		return
	}
	backend.counter.Wait(backend.ctx)
}

// Worker pool stop
func (backend *pool) Stop() {
	backend.workers = []WorkerInterface{}
	backend.cancel()
	close(backend.workerPool)
}

// Create simple jobs
func (backend *pool) NewJob(taskFn interface{}) (Job) {

	fn := reflect.ValueOf(taskFn)
	fnType := fn.Type()

	if fnType.Kind() != reflect.Func {
		panic("taskFn must be func")
	}

	return &jobFunc{
		fn:       fn,
		maxRetry: -1,
		pool:     backend,
	}
}

func (backend *pool) Insert(job jobCall) () {
	if backend.counter != nil {
		backend.counter.Increment()
	}
	backend.queue.Insert(job)
}

// apply factory to create queue, middlewares and others
func (backend *pool) Use(args ... interface{}) Pool {
	for _, arg := range args {
		switch arg.(type) {
		case CounterMiddleware:
			backend.prepareCounter = arg.(CounterMiddleware)
		case PoolQueue:
			backend.prepareQueue = arg.(PoolQueue)
		}
	}
	return backend
}

func (backend *pool) InsertPeriodic(job Job, period interface{}, arguments ... interface{}) (PeriodicJob, error) {
	var pJob PeriodicJob
	switch period.(type) {
	case time.Duration:
		pJob = NewTimeDurationJob(job, period.(time.Duration), arguments ...)
	case string:
		pJob = NewCronJob(job, period.(string), arguments ...)
	}
	if pJob == nil {
		return nil, errors.New("invalid period")
	}
	backend.periodicJob = append(backend.periodicJob, pJob)
	return pJob, nil
}

func (backend *pool) Context() context.Context {
	return backend.ctx
}

func New() Pool {
	return &pool{}
}

func processor(ctx context.Context, queue Queue, pool WorkerPoolType) {
	go func() {
		for {
			var job jobCall

			select {
			case <-ctx.Done():
				continue
			default:
				job = queue.Pop()
			}

			// if job is nil, then the context is closed
			if job == nil {
				continue
			}

			select {
			case <-ctx.Done():
				return
			case worker := <- pool:
				worker.AddJob(job)
			}
		}
	}()
}

// periodicProcessor is func for running periodic jobs
func periodicProcessor(ctx context.Context, periodicJobs []PeriodicJob) {
	go func() {
		lastCall := time.Now()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				maxInterval := lastCall.Add(time.Minute)
				queue := []NextJob{}
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
						queue = append(queue, NextJob{
							Job:  job,
							Next: next,
						})
					}
				}

				if len(queue) == 0 {
					select {
					case <-time.After(time.Second * 30):
						continue
					case <-ctx.Done():
						return
					}
				}

				// sort queue by time
				sort.Sort(NextJobSorter(queue))

				for _, pJob := range queue {
					select {
					case <-time.After(pJob.Next.Sub(lastCall)):
						lastCall = time.Now()
						pJob.Job.Run()
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
}
