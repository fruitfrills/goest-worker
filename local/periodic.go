package local

import (
	"github.com/gorhill/cronexpr"
	"time"
	goestworker "goest-worker"

)

type cronPeriodicJob struct {
	goestworker.PeriodicJob
	job 		goestworker.Job
	expr 		*cronexpr.Expression
	args		[]interface{}
}

func (pJob *cronPeriodicJob) Next(current time.Time) (time.Time){
	return pJob.expr.Next(current)
}

func (pJob *cronPeriodicJob) Run () (goestworker.JobInstance) {
	return pJob.job.Run(pJob.args ...)
}

type timeDurationPeriodicJob struct{
	goestworker.PeriodicJob
	job 		goestworker.Job
	last		time.Time
	duration 	time.Duration
	args		[]interface{}
}

func (pJob *timeDurationPeriodicJob) Next(current time.Time) (time.Time) {
	if current.Add(pJob.duration).Sub(pJob.last) < 0 {
		return current.Add(pJob.last.Sub(current)).Add(pJob.duration)
	}
	return current.Add(pJob.duration)
}

func (pJob *timeDurationPeriodicJob) Run () (goestworker.JobInstance) {
	pJob.last = time.Now()
	return pJob.job.Run(pJob.args ...)
}

type nextJob struct {
	job		goestworker.PeriodicJob
	next	time.Time
}
// for soi
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