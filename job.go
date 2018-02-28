package goest_worker

import (
	"reflect"
	"errors"
)

var ErrorJobDropped = errors.New("job is dropped")
var ErrorJobPanic = errors.New("job is panic")
var ErrorJobBind = errors.New("first argument is not job instance")

type Job interface {
	Run(args ... interface{}) JobInstance
	RunEvery(period interface{}, args ... interface{}) PeriodicJob
	Bind(bool) Job
	SetMaxRetry(int) Job
}


type JobInstance interface {
	Wait() JobInstance
	Result() ([]interface{}, error)

	getWaitChan() (chan JobInstance)
	call() JobInstance
	drop() JobInstance
}

type JobInjection interface {
	Retry() JobInstance
	WaitingFor(JobInstance) (JobInstance)
}

type jobFunc struct {
	Job

	// main func
	fn    			reflect.Value

	// dispatcher
	pool 			PoolInterface

	// self instance as first argument
	bind			bool

	maxRetry		int
}

type jobFuncInstance struct {
	JobInstance
	JobInjection
	// main jon
	job 			*jobFunc

	// argument for func
	args    		[]reflect.Value

	// result after calling func
	results 		[]reflect.Value

	// done channel for waiting
	done    		chan bool

	// waiting channel
	wait			chan JobInstance

	// for catching panic
	error			error

	// retry
	retry 			int
}

// create simple jobs
func NewJob(taskFn interface{}) (Job) {

	fn := reflect.ValueOf(taskFn)
	fnType := fn.Type()

	if fnType.Kind() != reflect.Func {
		panic("job is not func")
	}

	return &jobFunc{
		fn: fn,
		pool: MainPool,
		maxRetry: -1,
	}
}

func (job *jobFunc) SetMaxRetry (i int) (Job) {
	if (i < -1) {
		panic(`invalid count of retry`)
	}
	job.maxRetry = i
	return job
}

// calling func and close channel
func (jobInstance *jobFuncInstance) call() JobInstance {
	defer func() {
		// error handling
		if r := recover(); r != nil {
			var err error
			switch e := r.(type) {
			case string:
				err = errors.New(e)
			case error:
				err = e
			default:
				err = ErrorJobPanic
			}
			jobInstance.error = err
		}
		close(jobInstance.done)
		close(jobInstance.wait)
	}()
	jobInstance.results = jobInstance.job.fn.Call(jobInstance.args)
	return jobInstance
}

// open `done` channel and add task to queue of tasks
func (job *jobFunc) Run(arguments ... interface{}) (JobInstance) {
	in := make([]reflect.Value,  job.fn.Type().NumIn())
	for i, arg := range arguments {
		in[i] = reflect.ValueOf(arg)
	}
	instance := &jobFuncInstance{
		job: job,
		done: make(chan bool),
		wait: make(chan JobInstance),
		retry: job.maxRetry,
	}
	// if job.bind == true, set jobinstance as first argument
	if job.bind {
		in = append([]reflect.Value{reflect.ValueOf(instance)}, in[0:len(in)-1]...)
	}
	instance.args = in
	job.pool.addJobToPool(instance)
	return instance
}

// set bind
func (job *jobFunc) Bind(bind bool) (Job) {
	if job.fn.Type().In(0).Name() != "JobInjection" {
		panic(ErrorJobBind)
	}
	job.bind = bind
	return job
}

// run task every. arg may be string (cron like), time.Duration and time.time
func (job *jobFunc) RunEvery(period interface{}, arguments ... interface{}) (PeriodicJob) {
	return job.pool.addPeriodicJob(job, period, arguments ...)
}

// waiting tasks, call this after `Do`
func (jobInstance *jobFuncInstance) Wait() (JobInstance) {
	<-jobInstance.done
	return jobInstance
}

// dropping job
func (jobInstance *jobFuncInstance) drop () (JobInstance) {
	jobInstance.error = ErrorJobDropped
	jobInstance.done <- false
	return jobInstance
}

// get slice of results
func (jobInstance *jobFuncInstance) Result() ([]interface{}, error) {
	var result []interface{}
	for _, res := range jobInstance.results {
		result = append(result, res.Interface())
	}
	return result, jobInstance.error
}

func (jobInstance *jobFuncInstance) Retry() (JobInstance) {
	if jobInstance.retry == 0 {
		panic(`max retry`)
	}
	if jobInstance.retry != -1 {
		jobInstance.retry -=1
	}
	var in []interface{}
	for _, arg := range jobInstance.args{
		in = append(in, arg.Interface())
	}
	if jobInstance.job.bind{
		in = append([]interface{}{}, in[1:] ...)
	}
	return jobInstance.job.Run(in ...)
}

/**
 TODO: this method must be report to worker about state waiting for avoid dedlock

 */
func (jobInstance *jobFuncInstance) WaitingFor(job JobInstance) (JobInstance) {
	jobInstance.wait <- job
	job.Wait()
	return job
}

func (jobInstance *jobFuncInstance) getWaitChan() (chan JobInstance) {
	return jobInstance.wait
}