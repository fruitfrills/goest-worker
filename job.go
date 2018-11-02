package goest_worker

import (
	"reflect"
	"context"
	"errors"
)

// jobFunc is the wrapper around any function which can be execute in pool
type jobFunc struct {
	fn       reflect.Value
	pool     workerPool
	bind     bool
	maxRetry int
}

// jobFunc is the instance of running job at pool
// you can getting this in func if you use bind
type jobFuncInstance struct {
	job     *jobFunc
	args    []reflect.Value
	results []reflect.Value
	ctx     context.Context
	cancel  context.CancelFunc
	error   error
}

// Run - running function in any arguments
// creates JobInstance, context, parse args and put job to queue at pool
func (job *jobFunc) Run(arguments ... interface{}) (JobInstance) {
	in := make([]reflect.Value, job.fn.Type().NumIn())
	for i, arg := range arguments {
		in[i] = reflect.ValueOf(arg)
	}
	ctx, cancel := context.WithCancel(job.pool.Context())
	instance := &jobFuncInstance{
		job:    job,
		ctx:    ctx,
		cancel: cancel,
	}
	// if job.bind == true, set jobinstance as first argument
	if job.bind {
		in = append([]reflect.Value{reflect.ValueOf(instance)}, in[0:len(in)-1]...)
	}
	instance.args = in
	job.pool.AddJobToPool(instance)
	return instance
}

// run task every. arg may be string (cron like), time.Duration and time.time
func (job *jobFunc) RunEvery(period interface{}, arguments ... interface{}) (PeriodicJob, error) {
	return job.pool.AddPeriodicJob(job, period, arguments ...)
}

func (job *jobFunc) Bind(val bool) Job{
	job.bind = true
	return job
}

func (job *jobFunc) SetMaxRetry(val int) Job{
	job.maxRetry = val
	return job
}

func (jobInstance *jobFuncInstance) Wait() JobInstance{
	<- jobInstance.ctx.Done()
	return jobInstance
}

// Private method for calling func
func (jobInstance *jobFuncInstance) Call() {
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
				err = errors.New("unknown error")
			}
			jobInstance.error = err
			jobInstance.cancel()
		}
	}()
	select {
	case <- jobInstance.ctx.Done():
		return
	default:
		jobInstance.results = jobInstance.job.fn.Call(jobInstance.args)
		jobInstance.cancel()
	}
	return
}

func (jobInstance *jobFuncInstance) Cancel() {
	jobInstance.cancel()
}

func (jobInstance *jobFuncInstance) Context() context.Context {
	return jobInstance.ctx
}

// get slice of results
func (jobInstance *jobFuncInstance) Result() ([]interface{}, error) {
	var result []interface{}
	for _, res := range jobInstance.results {
		result = append(result, res.Interface())
	}
	return result, jobInstance.error
}

