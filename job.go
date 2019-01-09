package goest_worker

import (
	"reflect"
	"context"
	"github.com/pkg/errors"
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
	priority int
}

// Run - running function in any arguments
// creates JobInstance, context, parse args and put job to queue at pool
func (job *jobFunc) Run(arguments ... interface{}) (JobInstance) {

	// create context
	ctx, cancel := context.WithCancel(job.pool.Context())
	instance := &jobFuncInstance{
		job:    job,
		ctx:    ctx,
		cancel: cancel,
	}
	numIn := job.fn.Type().NumIn()
	lenArgs := len(arguments)
	if job.bind == true {
		lenArgs += 1
	}
	if numIn != lenArgs{
		instance.error = errors.New("invalid input parameter count")
		instance.cancel()
		return instance
	}
	in := make([]reflect.Value, numIn)

	for i, arg := range arguments {
		in[i] = reflect.ValueOf(arg)
	}

	// if job.bind == true, set jobinstance as first argument
	if job.bind {
		in = append([]reflect.Value{reflect.ValueOf(instance)}, in[0:len(in)-1]...)
	}
	instance.args = in
	job.pool.Insert(instance)
	return instance
}

func (job *jobFunc) RunWithPriority(priority int, arguments ... interface{}) (JobInstance) {
	in := make([]reflect.Value, job.fn.Type().NumIn())
	for i, arg := range arguments {
		in[i] = reflect.ValueOf(arg)
	}
	ctx, cancel := context.WithCancel(job.pool.Context())
	instance := &jobFuncInstance{
		job:    job,
		ctx:    ctx,
		cancel: cancel,
		priority: priority,
	}
	// if job.bind == true, set jobinstance as first argument
	if job.bind {
		in = append([]reflect.Value{reflect.ValueOf(instance)}, in[0:len(in)-1]...)
	}
	instance.args = in
	job.pool.Insert(instance)
	return instance
}

// run task every. arg may be string (cron like), time.Duration and time.time
func (job *jobFunc) RunEvery(period interface{}, arguments ... interface{}) (PeriodicJob, error) {
	return job.pool.InsertPeriodic(job, period, arguments ...)
}

func (job *jobFunc) Bind(val bool) Job{
	job.bind = true
	return job
}

// TODO: implement this!
func (job *jobFunc) SetMaxRetry(val int) Job{
	job.maxRetry = val
	return job
}

func (jobInstance *jobFuncInstance) Wait() JobInstance{
	<- jobInstance.ctx.Done()
	return jobInstance
}

// Get priority job
func (jobFuncInstance *jobFuncInstance) Priority() int {
	return jobFuncInstance.priority
}

// Private method for calling func
func (jobInstance *jobFuncInstance) call() {
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

// get results
func (jobInstance *jobFuncInstance) Results(args ... interface{}) error {

	// check context
	select {
	case <- jobInstance.ctx.Done():
		break
	default:
		return errors.New("task not completed")
	}

	// if has error
	if jobInstance.error != nil {
		return jobInstance.error
	}
	// check fn args
	if jobInstance.job.fn.Type().NumOut() != len(args) {
		return errors.Errorf("invalid output parameter count")
	}

	// prepare results
	for i := range args {
		if args[i] == nil {
			continue
		}
		val := reflect.ValueOf(args[i]).Elem()
		if val.Type() != jobInstance.results[i].Type() {
			return errors.Errorf("incorect type for %v %v", val.Type(), jobInstance.results[i].Type())
		}
		val.Set(jobInstance.results[i])
	}
	return nil
}