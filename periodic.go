package goest_worker

import (
	"github.com/gorhill/cronexpr"
	"time"
)

type cronPeriodicJob struct {
	job   Job
	expr  *cronexpr.Expression
	args  []interface{}
}

func (pJob *cronPeriodicJob) Next(current time.Time) (time.Time) {
	return pJob.expr.Next(current)
}

func (pJob *cronPeriodicJob) Run() () {
	pJob.job.Run(pJob.args ...)
	return
}

func NewTimeDurationJob(job Job, duration time.Duration, arguments ... interface{}) (PeriodicJob) {
	return  &timeDurationPeriodicJob{
		job:      job,
		duration: duration,
		last:     time.Now(),
		args: 	  arguments,
	}
}

func NewCronJob (job Job, expr string, arguments ... interface{}) (PeriodicJob) {
	return &cronPeriodicJob{
		job:  job,
		expr: cronexpr.MustParse(expr),
		args: arguments,
	}
}

type timeDurationPeriodicJob struct {
	job      Job
	last     time.Time
	duration time.Duration
	args     []interface{}
}

// time of next run
func (pJob *timeDurationPeriodicJob) Next(current time.Time) (time.Time) {
	if current.Add(pJob.duration).Sub(pJob.last) < 0 {
		return current.Add(pJob.last.Sub(current)).Add(pJob.duration)
	}
	return current.Add(pJob.duration)
}

func (pJob *timeDurationPeriodicJob) Run() () {
	pJob.last = time.Now()
	pJob.job.Run(pJob.args ...)
	return
}

// struct for sorting jobs by next
type nextJob struct {
	Job PeriodicJob
	Next time.Time
}

// for sort job by next
type nextJobSorter []nextJob

func (n nextJobSorter) Len() int {
	return len(n)
}

func (n nextJobSorter) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n nextJobSorter) Less(i, j int) bool {
	return n[i].Next.Sub(n[j].Next) < 0
}