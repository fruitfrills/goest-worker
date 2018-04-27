package goest_worker

import (
	"sync/atomic"
	"sync"
	"github.com/yobayob/goest-worker/common"
	"github.com/yobayob/goest-worker/local"
)

// pool states
const (
	POOL_STOPPED int32 = iota
	POOL_STARTED
)

// main wrapper over workers pool
type pool struct {

	// implements
	common.PoolInterface

	// state mutex
	sync.Mutex

	// pool's state
	state int32

	// backend
	backend 		common.PoolBackendInterface
}

// create pool of workers
func New(args ... common.PoolBackendInterface) (common.PoolInterface) {
	if len(args) > 1 {
		panic("invalid count args")
	}
	if len(args) == 1 {
		return &pool{
			backend: 		args[0],
			state:          POOL_STOPPED,
		}
	}
	return &pool{
		backend: 		&local.LocalBackend{},
		state:          POOL_STOPPED,
	}
}

// put job to queue
func (p *pool) AddJobToPool(task common.JobInstance) () {
	p.backend.AddJobToPool(p, task)
}

func (p *pool) AddPeriodicJob(job common.Job, period interface{}, arguments ... interface{}) (common.PeriodicJob) {
	return p.backend.AddPeriodicJob(p, job, period, arguments ...)
}

// check state
func (p *pool) IsStopped() (bool) {
	return atomic.LoadInt32(&(p.state)) == POOL_STOPPED
}

// set state
func (p *pool) setState(state int32) () {
	atomic.StoreInt32(&(p.state), int32(state))
}


// close all channels, stopping workers and periodic tasks
// if the dispatcher is restarted, all tasks will be stoppep. You need starts tasks again ...
func (p *pool) Stop() (common.PoolInterface) {
	p.Lock()
	defer func() {
		p.setState(POOL_STOPPED)
		p.Unlock()
	}()
	return p.backend.Stop(p)
}

func (p *pool) Start(count int) (common.PoolInterface) {
	p.Lock()
	defer func() {
		p.setState(POOL_STARTED)
		p.Unlock()
	}()
	return p.backend.Start(p, count)
}

// locker
func (p *pool) Lock() {
	p.Mutex.Lock()
}

// unlocker
func (p *pool) Unlock() {
	p.Mutex.Unlock()
}

// create job for current pool
func (p *pool) NewJob(taskFn interface{}) (job common.Job) {
	return p.backend.NewJob(p, taskFn)
}

// register job for current pool
func (p *pool) Register(name string, taskFn interface{}) (common.Job) {
	p.Lock()
	defer p.Unlock()
	return p.backend.Register(p, name, taskFn)
}