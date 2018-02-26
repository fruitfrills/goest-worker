package goest_worker

import (
	"reflect"
	"errors"
	"sync/atomic"
)

const (
	JOB_WAITING int32 = iota
	JOB_SUCCESS
	JOB_ERROR
	JOB_DROPPED
	JOB_PERIODIC
)

type Job interface {
	call()
	Run() Job
	RunEvery(interface{}) Job
	Wait() Job
	Result() []interface{}

	drop() Job
}

type jobFunc struct {
	Job
	fn    			reflect.Value
	args    		[]reflect.Value
	results 		[]reflect.Value
	done    		chan bool

	// job states
	// while without mutex
	state 			int32
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
		state: JOB_WAITING,
	}, nil
}

func (job *jobFunc) setState(state int32) () {
	atomic.StoreInt32(&(job.state), int32(state))
}

func (job *jobFunc) getState() (int32) {
	return atomic.LoadInt32(&(job.state))
}

// calling func and close channel
func (job *jobFunc) call() {

	defer func() {
		// error handling
		if r := recover(); r != nil {
			job.setState(JOB_ERROR)
		}
		close(job.done)
	}()

	job.results = job.fn.Call(job.args)
	if job.getState() != JOB_PERIODIC {
		job.setState(JOB_SUCCESS)
	}
}

// open `done` channel and add task to queue of tasks
func (task *jobFunc) Run() (Job) {
	task.done = make(chan bool)
	Pool.addTask(task)
	return task
}

// run task every. arg may be string (cron like), time.Duration and time.time
func (job *jobFunc) RunEvery(arg interface{}) (Job) {
	job.setState(JOB_PERIODIC)
	Pool.addTicker(job, arg)
	return job
}

// waiting tasks, call this after `Do`
func (job *jobFunc) Wait() (Job) {
	<-job.done
	return job
}

// dropping job
func (job *jobFunc) drop () (Job) {
	job.done <- false
	job.setState(JOB_DROPPED)
	return job
}

// get slice of results
func (job *jobFunc) Result() ([]interface{}) {
	var result []interface{}
	for _, res := range job.results {
		result = append(result, res.Interface())
	}
	return result
}
