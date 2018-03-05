package local

import (
	"reflect"
	"errors"
	"goest-worker/common"
)

type jobFunc struct {
	common.Job

	// main func
	fn    			reflect.Value

	// dispatcher
	pool 			common.PoolInterface

	// self instance as first argument
	bind			bool

	maxRetry		int
}

type jobFuncInstance struct {
	common.JobInstance
	common.JobInjection
	// main jon
	job 			*jobFunc

	// argument for func
	args    		[]reflect.Value

	// result after calling func
	results 		[]reflect.Value

	// done channel for waiting
	done    		chan bool


	// for catching panic
	error			error

	// retry
	retry 			int
}

func (job *jobFunc) SetMaxRetry (i int) (common.Job) {
	if (i < -1) {
		panic(`invalid count of retry`)
	}
	job.maxRetry = i
	return job
}


// open `done` channel and add task to queue of tasks
func (job *jobFunc) Run(arguments ... interface{}) (common.JobInstance) {
	in := make([]reflect.Value,  job.fn.Type().NumIn())
	for i, arg := range arguments {
		in[i] = reflect.ValueOf(arg)
	}
	instance := &jobFuncInstance{
		job: job,
		done: make(chan bool),
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
func (job *jobFunc) Bind(bind bool) (common.Job) {
	if job.fn.Type().In(0).Name() != "JobInjection" {
		panic(common.ErrorJobBind)
	}
	job.bind = bind
	return job
}

// run task every. arg may be string (cron like), time.Duration and time.time
func (job *jobFunc) RunEvery(period interface{}, arguments ... interface{}) (common.PeriodicJob) {
	return job.pool.AddPeriodicJob(job, period, arguments ...)
}

// calling func and close channel
func (jobInstance *jobFuncInstance) Call() common.JobInstance {
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
				err = common.ErrorJobPanic
			}
			jobInstance.error = err
		}
		close(jobInstance.done)
	}()
	jobInstance.results = jobInstance.job.fn.Call(jobInstance.args)
	return jobInstance
}

// waiting tasks, call this after `Do`
func (jobInstance *jobFuncInstance) Wait() (common.JobInstance) {
	<-jobInstance.done
	return jobInstance
}

// dropping job
func (jobInstance *jobFuncInstance) Drop () (common.JobInstance) {
	jobInstance.error = common.ErrorJobDropped
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

func (jobInstance *jobFuncInstance) Retry() (common.JobInstance) {
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
