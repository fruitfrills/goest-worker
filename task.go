package goest_worker

import (
	"time"
	"reflect"
	"errors"
)

type Task struct {
	task 		reflect.Value
	args 		[]reflect.Value
	results 	[]reflect.Value
	done 		chan bool
}

type PeriodicTask struct {
	Task
	Period time.Duration
}


/**
	task: function
	arguments: arguments for this function
 */
func NewTask(taskFn interface{}, arguments ... interface{}) (task *Task, err error) {

	fn := reflect.ValueOf(taskFn)
	fnType := fn.Type()

	if fnType.Kind() != reflect.Func{
		return nil, errors.New("Task is not func")

	}

	if fn.Type().NumIn() != len(arguments){
		return nil, errors.New("Invalid num arguments")
	}

	in := make([]reflect.Value, fn.Type().NumIn())

	for i, arg := range arguments{
		in[i] = reflect.ValueOf(arg)
	}

	return &Task{
		task: fn,
		args: in,
		done: make(chan bool),
	}, nil
}

func (task *Task) Call() {
	defer close(task.done)
	task.results = task.task.Call(task.args)
}

func (task *Task) Wait () {
	Pool.AddTask(task)
	<- task.done
}

func (task *Task) Result() ([]interface{}) {
	var result []interface{}
	for _, res := range task.results {
		result = append(result, res.Interface())
	}
	return result
}