package local

import (
	"testing"
	"goest-worker/common"
	"time"
)

type mockPool struct {
	stop bool
}

func (p *mockPool) Lock () {
	return
}

func (p *mockPool) Unlock () {
	return
}

func (p *mockPool) IsStopped () bool {
	return p.stop
}

func (p *mockPool) Start(count int) (common.PoolInterface) {
	return p
}

func (p *mockPool) Stop() (common.PoolInterface) {
	return p
}

func (p *mockPool) NewJob(taskFn interface{}) (common.Job){
	return nil
}

func (p *mockPool) Register(name string, taskFn interface{}) common.Job {
	return nil
}

func (p *mockPool) AddJobToPool (job common.JobInstance) {
	return
}

func (p *mockPool) AddPeriodicJob (job common.Job, period interface{}, arguments ... interface{}) (common.PeriodicJob) {
	return nil
}

type mockJob struct {
}

func (job *mockJob) SetMaxRetry (i int) (common.Job) {
	return job
}


func (job *mockJob) Run(arguments ... interface{}) (common.JobInstance) {
	return nil
}

func (job *mockJob) Bind(bind bool) (common.Job) {
	return job
}

func (job *mockJob) RunEvery(period interface{}, arguments ... interface{}) (common.PeriodicJob) {
	return nil
}


func TestLocalBackend_AddPeriodicJob(t *testing.T) {
	defer func() {
		if r := recover(); r != common.ErrorJobAdding{
			t.Errorf(`adding tasks on runned pool`)
		}
	}()
	p := &mockPool{stop: false,}
	backend := LocalBackend{}
	backend.Start(p, 10)
	backend.AddPeriodicJob(p, &mockJob{}, time.Second,)
}

func TestLocalBackend_AddPeriodicJob2(t *testing.T) {
	p := &mockPool{stop: true,}
	backend := LocalBackend{}
	backend.AddPeriodicJob(p, &mockJob{}, time.Second,)

	if len(backend.periodicJob) != 1 {
		t.Errorf("periodic job is not added")
	}
}

