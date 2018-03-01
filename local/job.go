package local

import (
	"reflect"
	"errors"
	goestworker "goest-worker/common"
)

type jobFunc struct {
	goestworker.Job

	// main func
	fn    			reflect.Value

	// dispatcher
	pool 			goestworker.PoolInterface

	// self instance as first argument
	bind			bool

	maxRetry		int
}

type jobFuncInstance struct {
	goestworker.JobInstance
	goestworker.JobInjection
	// main jon
	job 			*jobFunc

	// argument for func
	args    		[]reflect.Value

	// result after calling func
	results 		[]reflect.Value

	// done channel for waiting
	done    		chan bool

	// waiting channel
	wait			chan goestworker.JobInstance

	// for catching panic
	error			error

	// retry
	retry 			int
}

func (job *jobFunc) SetPool (p goestworker.PoolInterface) (goestworker.Job) {
	job.pool = p
	return job
}

func (job *jobFunc) SetMaxRetry (i int) (goestworker.Job) {
	if (i < -1) {
		panic(`invalid count of retry`)
	}
	job.maxRetry = i
	return job
}

// calling func and close channel
func (jobInstance *jobFuncInstance) Call() goestworker.JobInstance {
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
				err = goestworker.ErrorJobPanic
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
func (job *jobFunc) Run(arguments ... interface{}) (goestworker.JobInstance) {
	in := make([]reflect.Value,  job.fn.Type().NumIn())
	for i, arg := range arguments {
		in[i] = reflect.ValueOf(arg)
	}
	instance := &jobFuncInstance{
		job: job,
		done: make(chan bool),
		wait: make(chan goestworker.JobInstance),
		retry: job.maxRetry,
	}
	// if job.bind == true, set jobinstance as first argument
	if job.bind {
		in = append([]reflect.Value{reflect.ValueOf(instance)}, in[0:len(in)-1]...)
	}
	instance.args = in
	job.pool.AddJobToPool(instance)
	return instance
}

// set bind
func (job *jobFunc) Bind(bind bool) (goestworker.Job) {
	if job.fn.Type().In(0).Name() != "JobInjection" {
		panic(goestworker.ErrorJobBind)
	}
	job.bind = bind
	return job
}

// run task every. arg may be string (cron like), time.Duration and time.time
func (job *jobFunc) RunEvery(period interface{}, arguments ... interface{}) (goestworker.PeriodicJob) {
	return job.pool.AddPeriodicJob(job, period, arguments ...)
}

// waiting tasks, call this after `Do`
func (jobInstance *jobFuncInstance) Wait() (goestworker.JobInstance) {
	<-jobInstance.done
	return jobInstance
}

// dropping job
func (jobInstance *jobFuncInstance) Drop () (goestworker.JobInstance) {
	jobInstance.error = goestworker.ErrorJobDropped
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

func (jobInstance *jobFuncInstance) Retry() (goestworker.JobInstance) {
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
