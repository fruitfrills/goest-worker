package local

import (
	"github.com/gorhill/cronexpr"
	"time"
	"goest-worker/common"

	"sync"
	"sync/atomic"
)

const (
	job_free int32 = iota
	job_busy
)

type cronPeriodicJob struct {
	common.PeriodicJob

	sync.Mutex
	state		int32
	job 		common.Job
	expr 		*cronexpr.Expression
	args		[]interface{}
}

func (pJob *cronPeriodicJob) Next(current time.Time) (time.Time){
	return pJob.expr.Next(current)
}

func (pJob *cronPeriodicJob) Run () () {
	pJob.Lock()
	defer pJob.Unlock()
	if pJob.isBusy(){
		// log.Warning
		return
	}
	pJob.setState(job_busy)
	jobInstance := pJob.job.Run(pJob.args ...)
	go func() {
		jobInstance.Wait()
		pJob.Lock()
		pJob.setState(job_free)
		pJob.Unlock()
	}()
	return
}

func (pJob *cronPeriodicJob) setState(state int32) {
	atomic.StoreInt32( &(pJob.state), state)
}

func (pJob *cronPeriodicJob) isBusy() bool {
	return atomic.LoadInt32(&(pJob.state)) == job_busy
}

type timeDurationPeriodicJob struct{
	common.PeriodicJob

	sync.Mutex
	state		int32
	job 		common.Job
	last		time.Time
	duration 	time.Duration
	args		[]interface{}
}

// time of next run
func (pJob *timeDurationPeriodicJob) Next(current time.Time) (time.Time) {
	if current.Add(pJob.duration).Sub(pJob.last) < 0 {
		return current.Add(pJob.last.Sub(current)).Add(pJob.duration)
	}
	return current.Add(pJob.duration)
}

func (pJob *timeDurationPeriodicJob) Run () () {
	pJob.Lock()
	defer pJob.Unlock()

	// run once instance
	if pJob.isBusy(){
		// log.Warning
		return
	}
	pJob.setState(job_busy)
	pJob.last = time.Now()
	jobInstance := pJob.job.Run(pJob.args ...)

	go func() {
		jobInstance.Wait()
		pJob.Lock()
		pJob.setState(job_free)
		pJob.Unlock()
	}()

	return
}

// set state to periodic job
func (pJob *timeDurationPeriodicJob) setState(state int32) {
	atomic.StoreInt32( &(pJob.state), state)
}

// check busy job
func (pJob *timeDurationPeriodicJob) isBusy() bool {
	return atomic.LoadInt32(&(pJob.state)) == job_busy
}

// struct for sorting jobs by next
type nextJob struct {
	job		common.PeriodicJob
	next	time.Time
}
// for sort job by next
type nextJobSorter []nextJob

func (n nextJobSorter) Len () int {
	return len(n)
}

func (n nextJobSorter) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n nextJobSorter) Less(i, j int) bool {
	return n[i].next.Sub(n[j].next) < 0
}