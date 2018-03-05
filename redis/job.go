package redis

import (
	"goest-worker/common"
	"reflect"
	"encoding/json"
)

type jobRedisFunc struct {
	common.Job

	// main func
	fn    			reflect.Value

	// dispatcher
	pool 			common.PoolInterface

	// self instance as first argument
	bind			bool

	// max count of retry
	maxRetry		int

	// name of func
	name			string

	// register on unregister job
	isRegister		bool
}

type jobRedisFuncInstance struct {
	common.JobInstance
	common.JobInjection
	// main jon
	job 			*jobRedisFunc

	// argument for func
	args    		[]reflect.Value

	// result after calling func
	results 		[]reflect.Value

	// done channel for waiting
	done    		chan bool

	// waiting channel
	wait			chan common.JobInstance

	// for catching panic
	error			error

	// retry
	retry 			int
}

// parse job
func Unmarshall(value string) (common.JobInstance, error){
	var job jobRedisFuncInstance
	err := json.Unmarshal([]byte(value), &job)
	return &job, err
}

func (job *jobRedisFunc) Run(args ... interface{}) (jobInstance common.JobInstance) {
	panic(`not implemented`)
	return
}

func (job *jobRedisFunc) RunEvery(period interface{}, args ... interface{}) (jobInstance common.JobInstance) {
	panic(`not implemented`)
	return
}

func (job *jobRedisFunc) Bind(bool) common.Job {
	return job
}

func (job *jobRedisFunc) SetMaxRetry(int) common.Job {
	return job
}
