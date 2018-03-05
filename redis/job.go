package redis

import (
	"goest-worker/common"
	"reflect"
	"encoding/json"
	db "github.com/go-redis/redis"

	"log"
	"errors"
	"math"
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

	JobName			string

	PoolName		string
}

type jobRedisFuncInstance struct {
	// main jon
	job 			*jobRedisFunc

	// argument for func
	Args    		[]interface{}

	// result after calling func
	results 		[]reflect.Value

	// done channel for waiting
	done    		chan bool

	// done channel for waiting
	pubSubChannel    *db.PubSub

	// name for job
	JobName			string

	// connect for pub sub
	conn 			*db.Client

	PoolName		string

	// pubsub name
	PubName			string

	// for catching panic
	error			error

	// register on unregister job
	IsRegister		bool

	// retry
	retry 			int
}


func (job *jobRedisFunc) Run(arguments ... interface{}) (jobInstance common.JobInstance) {
	instance := &jobRedisFuncInstance{
		job: job,
		retry: job.maxRetry,
		done: make(chan bool),
		PoolName: job.PoolName,
		JobName: job.name,
	}

	instance.Args = arguments
	job.pool.AddJobToPool(instance)

	// waiting and done
	go func() {
		var results jobResults
		msg, err := instance.pubSubChannel.ReceiveMessage()
		if err != nil{
			log.Println(err.Error())
			return
		}
		json.Unmarshal([]byte(msg.Payload), &results)
		if results.Err != ""{
			instance.error = errors.New(results.Err)
			instance.done <- true
			return
		}
		for i, res := range results.Results{
			instance.results = append(instance.results, replacerOut(job.fn, i, res))
		}
		instance.done <- true
	}()
	return instance
}

func (job *jobRedisFunc) RunEvery(period interface{}, arguments ... interface{}) (common.PeriodicJob) {
	return job.pool.AddPeriodicJob(job, period, arguments ...)
}

func (job *jobRedisFunc) Bind(bind bool) (common.Job) {
	if job.fn.Type().In(0).Name() != "JobInjection" {
		panic(common.ErrorJobBind)
	}
	job.bind = bind
	return job
}

func (job *jobRedisFunc) SetMaxRetry (i int) (common.Job) {
	if (i < -1) {
		panic(`invalid count of retry`)
	}
	job.maxRetry = i
	return job
}

// waiting tasks, call this after `Do`
func (jobInstance *jobRedisFuncInstance) Wait() (common.JobInstance) {
	<-jobInstance.done
	return jobInstance
}

// dropping job
func (jobInstance *jobRedisFuncInstance) Drop () (common.JobInstance) {
	jobInstance.error = common.ErrorJobDropped
	jobInstance.done <- false
	return jobInstance
}

type jobResults struct {
	Err string
	Results []interface{}
}

// calling func and close channel
func (jobInstance *jobRedisFuncInstance) Call() (common.JobInstance) {
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
			jobInstance.conn.Publish(jobInstance.PubName, jobResults{
				Err: err.Error(),
			})
			log.Println(err.Error())
		}
	}()
	var in []reflect.Value

	for i, arg := range jobInstance.Args {
		if jobInstance.job.bind {
			i +=1
		}
		in = append(in, replacerIn(jobInstance.job.fn, i, arg))
	}
	if jobInstance.job.bind {
		in = append([]reflect.Value{reflect.ValueOf(jobInstance)}, in...)
	}
	results := jobInstance.job.fn.Call(in)
	var resInterface []interface{}
	for _, res := range results {
		resInterface = append(resInterface, res.Interface())
	}

	resJson, err := json.Marshal(jobResults{
		Results: resInterface,
	})
	if err != nil{
		panic(err)
	}
	jobInstance.conn.Publish(jobInstance.PubName, resJson)
	return jobInstance
}

// get slice of results
func (jobInstance *jobRedisFuncInstance) Result() ([]interface{}, error) {
	var result []interface{}
	for _, res := range jobInstance.results {
		result = append(result, res.Interface())
	}
	return result, jobInstance.error
}

// get slice of results
func (jobInstance *jobRedisFuncInstance) Retry() (job common.JobInstance) {
	panic(`not implement`)
	return
}

func round(f float64) int {
	n1, n2 := math.Modf(f) // splits the float into components

	n := int(n1) // grab the integer part of the float
	if n2 > .5 { // and if the floating part is greater than .5
		n++ //  round it up!
	}

	return (n)
}

func uround(f float64) uint {
	n1, n2 := math.Modf(f) // splits the float into components

	n := uint(n1) // grab the integer part of the float
	if n2 > .5 { // and if the floating part is greater than .5
		n++ //  round it up!
	}

	return (n)
}

func replacerIn(fn reflect.Value, i int, arg interface{}) (reflect.Value) {
	// todo: fix this
	switch fn.Type().In(i).Kind() {
	case reflect.Int:
		return reflect.ValueOf(round(arg.(float64)))
	case reflect.Int8:
		return reflect.ValueOf(int8(round(arg.(float64))))
	case reflect.Int16:
		return reflect.ValueOf(int16(round(arg.(float64))))
	case reflect.Int32:
		return reflect.ValueOf(int32(round(arg.(float64))))
	case reflect.Uint:
		return reflect.ValueOf(uround(arg.(float64)))
	case reflect.Uint8:
		return reflect.ValueOf(uint8(uround(arg.(float64))))
	case reflect.Uint16:
		return reflect.ValueOf(uint16(uround(arg.(float64))))
	case reflect.Uint32:
		return reflect.ValueOf(uint32(uround(arg.(float64))))
	case reflect.Float32:
		return reflect.ValueOf(arg.(float32))
	case reflect.Float64:
		return reflect.ValueOf(arg.(float64))
	default:
		return reflect.ValueOf(arg)
	}
}

func replacerOut(fn reflect.Value, i int, arg interface{})(reflect.Value) {
	// todo: fix this
	switch fn.Type().Out(i).Kind() {
	case reflect.Int:
		return reflect.ValueOf(round(arg.(float64)))
	case reflect.Int8:
		return reflect.ValueOf(int8(round(arg.(float64))))
	case reflect.Int16:
		return reflect.ValueOf(int16(round(arg.(float64))))
	case reflect.Int32:
		return reflect.ValueOf(int32(round(arg.(float64))))
	case reflect.Uint:
		return reflect.ValueOf(uround(arg.(float64)))
	case reflect.Uint8:
		return reflect.ValueOf(uint8(uround(arg.(float64))))
	case reflect.Uint16:
		return reflect.ValueOf(uint16(uround(arg.(float64))))
	case reflect.Uint32:
		return reflect.ValueOf(uint32(uround(arg.(float64))))
	case reflect.Float32:
		return reflect.ValueOf(arg.(float32))
	case reflect.Float64:
		return reflect.ValueOf(arg.(float64))
	default:
		return reflect.ValueOf(arg)
	}
}