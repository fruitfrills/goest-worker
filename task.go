package goest_worker

import (
	"time"
	"reflect"
	"errors"
)

type Task struct {
	Task reflect.Value
	Args []reflect.Value
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
		Task: fn,
		Args: in,
	}, nil
}
