package goest_worker


import (
	"time"
)


type Job interface {

	// run job
	Run(args ... interface{}) JobInstance

	// create periodic job
	RunEvery(period interface{}, args ... interface{}) PeriodicJob

	// use job instance as first argument
	Bind(bool) Job

	// count max retry
	SetMaxRetry(int) Job
}

type JobInstance interface {

	// waitig results
	Wait() JobInstance

	// get results
	Result() ([]interface{}, error)

	// call
	Call() JobInstance

	// drop
	Drop() JobInstance
}

type JobInjection interface {
	// repeat job
	Retry() JobInstance
}

type PeriodicJob interface {
	Next(time.Time) time.Time
	Run() ()
}
