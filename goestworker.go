package goest_worker

import (
	"sync"
	"context"
	"time"
	"sort"
	"reflect"
	"errors"
)

// GoestWorker local worker's pool dispatcher
type goestWorker struct {

	// Inner context for worker's pool
	ctx         context.Context

	// Cancel context func
	cancel      context.CancelFunc

	// Chan for free workers
	workerPool  chan WorkerInterface

	// Chan for job
	jobQueue    chan jobCall

	// Periodic jobs
	periodicJob []PeriodicJob


	wg sync.WaitGroup
}

// Run worker pool
func (backend *goestWorker) Start(ctx context.Context, count int) GoestWorker {

	// create new context
	ctx, cancel := context.WithCancel(ctx)
	backend.ctx = ctx
	backend.cancel = cancel
	backend.jobQueue = make(chan jobCall)
	backend.workerPool = make(WorkerPoolType, count)

	// create workers
	for i := 0; i < count; i++ {
		worker := newWorker(backend.workerPool, &backend.wg)
		worker.Start(ctx)
	}

	// run processor
	proccessor(ctx, backend.jobQueue, backend.workerPool)

	// run periodic processor
	if len(backend.periodicJob) != 0 {
		periodicProcessor(ctx, backend.periodicJob)
	}
	return backend
}

func (backend *goestWorker) wait(ctx context.Context, wg *sync.WaitGroup) {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		done <- struct{}{}
	}()
	select {
	case <-done:
		break
	case <-ctx.Done():
		break
	}
	return
}

func (backend *goestWorker) Done() {
	backend.wg.Done()
}

func (backend *goestWorker) Wait() {
	backend.wait(backend.ctx, &backend.wg)
}

// Worker pool stop
func (backend *goestWorker) Stop() {
	backend.cancel()
	close(backend.jobQueue)
	close(backend.workerPool)
}


// Create simple jobs
func (backend *goestWorker) NewJob(taskFn interface{}) (Job) {

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

func (backend *goestWorker) AddJobToPool(job jobCall) () {
	backend.wg.Add(1)
	go func() {
		select {
		case <-backend.ctx.Done():
			return
		default:
			backend.jobQueue <- job
		}
	}()
}

func (backend *goestWorker) AddPeriodicJob(job Job, period interface{}, arguments ... interface{}) (PeriodicJob, error) {
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

func (backend *goestWorker) Context() context.Context {
	return backend.ctx
}

func New() GoestWorker {
	return &goestWorker{}
}

func proccessor(ctx context.Context, jobQueue chan jobCall, pool WorkerPoolType) {
	go func() {
		for {
			var job jobCall
			select {

			case job = <-jobQueue:

				if job == nil {
					return
				}
				break
			case <-ctx.Done():
				return
			}

			select {
			case <-ctx.Done():
				return
			case worker := <-pool:
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
