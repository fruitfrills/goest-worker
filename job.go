package goest_worker

import (
	"reflect"
	"errors"
)

var ErrorJobDropped = errors.New("job is dropped")
var ErrorJobPanic = errors.New("job is panic")

type Job interface {
	Run() Job
	RunEvery(interface{}) Job
	Wait() Job
	Result() ([]interface{}, error)

	call() Job
	drop() Job
}

type jobFunc struct {
	Job

	// main func
	fn    			reflect.Value

	// argument for func
	args    		[]reflect.Value

	// result after calling func
	results 		[]reflect.Value

	// done channel for waiting
	done    		chan bool

	// for catching panic
	error			error

	// dispatcher
	pool 			PoolInterface
}

// create simple jobs
func NewJob(taskFn interface{}, arguments ... interface{}) (task Job, err error) {

	fn := reflect.ValueOf(taskFn)
	fnType := fn.Type()

	if fnType.Kind() != reflect.Func {
		return nil, errors.New("job is not func")
	}

	if fn.Type().NumIn() != len(arguments) {
		return nil, errors.New("invalid num arguments")
	}

	in := make([]reflect.Value, fn.Type().NumIn())

	for i, arg := range arguments {
		in[i] = reflect.ValueOf(arg)
	}

	return &jobFunc{
		fn: fn,
		args: in,
		pool: MainPool,
	}, nil
}

// calling func and close channel
func (job *jobFunc) call() Job {
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
			job.error = err
		}
		close(job.done)
	}()
	job.results = job.fn.Call(job.args)
	return job
}

// open `done` channel and add task to queue of tasks
func (job *jobFunc) Run() (Job) {
	job.done = make(chan bool)
	job.pool.AddTask(job)
	return job
}

// run task every. arg may be string (cron like), time.Duration and time.time
func (job *jobFunc) RunEvery(arg interface{}) (Job) {
	job.pool.addTicker(job, arg)
	return job
}

// waiting tasks, call this after `Do`
func (job *jobFunc) Wait() (Job) {
	<-job.done
	return job
}

// dropping job
func (job *jobFunc) drop () (Job) {
	job.error = ErrorJobDropped
	job.done <- false
	return job
}

// get slice of results
func (job *jobFunc) Result() ([]interface{}, error) {
	var result []interface{}
	for _, res := range job.results {
		result = append(result, res.Interface())
	}
	return result, job.error
}
