package goest_worker

import (
	"github.com/gorhill/cronexpr"
	"time"
)

type PeriodicJob interface {
	Next(time.Time) time.Time
	run() (JobInstance)
}

type cronPeriodicJob struct {
	PeriodicJob
	job 		Job
	expr 		*cronexpr.Expression
	args		[]interface{}
}

func (pJob *cronPeriodicJob) Next(current time.Time) (time.Time){
	return pJob.expr.Next(current)
}

func (pJob *cronPeriodicJob) run () (JobInstance) {
	return pJob.job.Run(pJob.args ...)
}

type timeDurationPeriodicJob struct{
	PeriodicJob
	job 		Job
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

func (pJob *timeDurationPeriodicJob) run () (JobInstance) {
	pJob.last = time.Now()
	return pJob.job.Run(pJob.args ...)
}

type nextJob struct {
	job		PeriodicJob
	next	time.Time
}

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