package goest_worker

import (
	"context"
	"time"
	"sort"
	"reflect"
	"errors"
	"sync/atomic"
)

// Pool send to queue job from client
type pool struct {

	// Inner context for worker's pool
	ctx         context.Context

	// Cancel context func
	cancel      context.CancelFunc

	// Chan for free workers
	workerPool  chan WorkerInterface

	// Chan for job
	sendJobQueue    chan jobCall

	// Chan for job
	recieveJobQueue    chan jobCall

	// Periodic jobs
	periodicJob []PeriodicJob

	// Waiting group for waiting of pool finish all jobs
	jobsCounter uint64
}

// Run worker pool
func (backend *pool) Start(ctx context.Context, count int) Pool {

	// create new context
	ctx, cancel := context.WithCancel(ctx)
	backend.ctx = ctx
	backend.cancel = cancel

	// make unlimited queue for jobs
	send := make(chan jobCall, 1)
	receive := make(chan jobCall, count)
	proccessQueue(ctx, send, receive)

	backend.sendJobQueue = send
	backend.recieveJobQueue = receive
	backend.workerPool = make(WorkerPoolType, count)
	atomic.StoreUint64(&backend.jobsCounter, 0)
	// create workers
	for i := 0; i < count; i++ {
		worker := newWorker(backend.workerPool, &backend.jobsCounter)
		worker.Start(ctx)
	}

	// run processor
	proccessor(ctx, backend.recieveJobQueue, backend.workerPool)

	// run periodic processor
	if len(backend.periodicJob) != 0 {
		periodicProcessor(ctx, backend.periodicJob)
	}
	return backend
}

func proccessQueue(ctx context.Context, send chan jobCall, receive chan jobCall) {

	var exit = func() {
		close(send)
		close(receive)
	}

	go func() {
		heap := newJobHeap()
		for {
			var top *jobHeapNode
			select {
			case <-ctx.Done():
				exit()
				return
			default:
				top = heap.Top()
			}

			if top == nil {

				var job jobCall
				select {
				case <- ctx.Done():
					exit()
					return
				case job = <- send:
					break
				}

				// if queue empty try send job to receiver
				select {
				case <- ctx.Done():
					exit()
					return
				case receive <- job:
					continue
				default:
					heap.Insert(job)
				}

				continue
			}

			select {
			case <- ctx.Done():
				exit()
				return
			case receive <- top.Job():
				heap.Remove(top)
			case value := <- send:
				heap.Insert(value)
			}

		}
	}()
}

func (backend *pool) wait(ctx context.Context, counter *uint64) {
	for {
		select {
		case <- ctx.Done():
			return
		case <- time.After(time.Millisecond):
			if atomic.LoadUint64(counter) == 0 {
				return
			}
		}
	}
	return
}

func (backend *pool) Wait() {
	backend.wait(backend.ctx, &backend.jobsCounter)
}

// Worker pool stop
func (backend *pool) Stop() {
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

func (backend *pool) AddJobToPool(job jobCall) () {
	atomic.AddUint64(&backend.jobsCounter, 1)
	select {
	case <-backend.ctx.Done():
		return
	case backend.sendJobQueue <- job:
		return
	}
}

func (backend *pool) AddPeriodicJob(job Job, period interface{}, arguments ... interface{}) (PeriodicJob, error) {
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
