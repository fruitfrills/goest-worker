package goest_worker

import (
	"reflect"
	"errors"
)


type Job interface {
	call()
	Run() Job
	RunEvery(interface{}) Job
	Wait() Job
	Result() []interface{}
}

type jobFunc struct {
	Job
	fn    			reflect.Value
	args    		[]reflect.Value
	results 		[]reflect.Value
	done    		chan bool
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
	}, nil
}

// calling func and close channel
func (job *jobFunc) call() {
	defer close(job.done)
	job.results = job.fn.Call(job.args)
}

// open `done` channel and add task to queue of tasks
func (task *jobFunc) Run() (Job) {
	task.done = make(chan bool)
	Pool.addTask(task)
	return task
}

// run task every. arg may be string (cron like), time.Duration and time.time
func (task *jobFunc) RunEvery(arg interface{}) (Job) {
	Pool.addTicker(task, arg)
	return task
}

// waiting tasks, call this after `Do`
func (task *jobFunc) Wait() (Job) {
	<-task.done
	return task
}

// get slice of results
func (task *jobFunc) Result() ([]interface{}) {
	var result []interface{}
	for _, res := range task.results {
		result = append(result, res.Interface())
	}
	return result
}
