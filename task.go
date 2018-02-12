package goest_worker

import (
	"reflect"
	"errors"
	"time"
)

type Schedule struct{
	Weekday 	time.Weekday
	Hour 		int
	Minute		int
	Second		int
	NanoSecond  int
}

func (s *Schedule) Next() (time.Duration) {
	now := time.Now()
	next := now
	if (now.Weekday() != s.Weekday) && s.Weekday != 0 {
		next = next.AddDate(0,0, (7 - int((now.Weekday()))) + int(s.Weekday) )
	}
	next = time.Date(next.Year(), next.Month(), next.Day(), s.Hour, s.Minute, s.Second, s.NanoSecond, now.Location())
	diff := next.Sub(now)
	if diff < 0 {
		diff += time.Hour * 24
	}
	return diff
}

type TaskInterface interface {
	call()
	Run() TaskInterface
	Wait() TaskInterface
	Result() []interface{}
	RunEvery(interface{}) TaskInterface
}


type Task struct {
	TaskInterface
	task 		reflect.Value
	args 		[]reflect.Value
	results 	[]reflect.Value
	done 		chan bool
}

/**
	task: function
	arguments: arguments for this function
 */
func NewTask(taskFn interface{}, arguments ... interface{}) (task TaskInterface, err error) {

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
	}, nil
}

func (task *Task) call() {
	defer close(task.done)
	task.results = task.task.Call(task.args)
}


func (task *Task) RunEvery(arg interface{}) (TaskInterface) {
	Pool.addTicker(task, arg)
	return task
}

func (task *Task) Run () (TaskInterface){
	task.done = make(chan bool)
	Pool.AddTask(task)
	return task
}

func (task *Task) Wait () (TaskInterface) {
	<- task.done
	return task
}

func (task *Task) Result() ([]interface{}) {
	var result []interface{}
	for _, res := range task.results {
		result = append(result, res.Interface())
	}
	return result
}