package redis

import (
	"goest-worker/common"
	db "github.com/go-redis/redis"
	"time"
	"log"
	"encoding/json"
	"reflect"
	"runtime"
	"sync"
)

// local backend
type RedisBackend struct {
	common.PoolBackendInterface

	// can getting free worker from this chan
	workerPool chan common.WorkerInterface

	// slice of quit channel for soft finish
	workersPoolQuitChan []chan bool

	// periodic jobs
	periodicJob []common.PeriodicJob

	// redis queue name
	QueueName	string

	// redis addr
	Addr		string

	// redis pass
	Password	string

	// redis db num
	DbNum		int

	// connect to redis
	conn		*db.Client

	// map for register job
	registerJob		sync.Map

	// map for unregisterjob
	// run once and dropped
	unregisterJob 	sync.Map
}


func (backend *RedisBackend) AddJobToPool(p common.PoolInterface,job common.JobInstance) () {
	value, err := json.Marshal(job)
	if err != nil {
		panic(err)
	}
	backend.conn.RPush(backend.QueueName, value)
	return
}

func (backend *RedisBackend) AddPeriodicJob(common.PoolInterface, common.Job, interface{}, ... interface{}) (job common.PeriodicJob) {
	panic(`not implemented`)
	return
}

func (backend *RedisBackend) Scheduler(common.PoolInterface) {
	panic(`not implemented`)
	return
}

func (backend *RedisBackend) Processor(common.PoolInterface) {
	quitChan := make(chan bool)
	backend.workersPoolQuitChan = append(backend.workersPoolQuitChan, quitChan)
	for {
		select {
		case <-quitChan:
			return

		default:
			val := backend.conn.BLPop(time.Second, backend.QueueName)

			var err error

			if err = val.Err(); err != nil &&  err != db.Nil {
				log.Println(`Redis error`, err)
				continue
			}
			var job common.JobInstance
			if job, err = Unmarshall(val.Val()[1]); err != nil {
				log.Println(`Redis unmarshall error`, err)
				continue
			}
			var worker common.WorkerInterface
			// get free worker and send task
			worker = <-backend.workerPool
			worker.AddJob(job)
		}
	}
	return
}

func (backend *RedisBackend) Stop(common.PoolInterface) (p common.PoolInterface) {
	backend.conn.Close()

	// send close to all quit channels
	for _, quit := range backend.workersPoolQuitChan {
		close(quit)
	}

	backend.workersPoolQuitChan = [](chan bool){};
	close(backend.workerPool)
	return p
}

func (backend *RedisBackend) Start(p common.PoolInterface, count int) (common.PoolInterface) {
	// if dispatcher started - do nothing
	if !p.IsStopped() {
		return p
	}

	// create connect

	backend.conn = db.NewClient(&db.Options{
		Addr:     backend.Addr,
		Password: backend.Password,
		DB:       backend.DbNum,
	})

	for i := 0; i < count; i++ {
		worker := NewWorker(backend.workerPool)
		backend.workersPoolQuitChan = append(backend.workersPoolQuitChan, worker.GetQuitChan())
		worker.Start()
	}
	// main process
	go backend.Processor(p)

	// periodic proccess
	if  len(backend.periodicJob) != 0 {
		go backend.Scheduler(p)
	}

	return p
}

func (backend *RedisBackend) NewJob(p common.PoolInterface, taskFn interface{}) (job common.Job) {

	fn := reflect.ValueOf(taskFn)
	fnType := fn.Type()

	if fnType.Kind() != reflect.Func {
		panic("job is not func")
	}
	name := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	job = &jobRedisFunc{
		name: name,
		fn: fn,
		maxRetry: -1,
		pool: p,
	}
	backend.unregisterJob.Store(name, job)
	return job
}

func (backend *RedisBackend) Register(p common.PoolInterface, name string, taskFn interface{}) (job common.Job) {
	fn := reflect.ValueOf(taskFn)
	fnType := fn.Type()

	if fnType.Kind() != reflect.Func {
		panic("job is not func")
	}
	job = &jobRedisFunc{
		name: name,
		fn: fn,
		maxRetry: -1,
		pool: p,
		isRegister: true,
	}
	backend.registerJob.Store(name, job)
	return job
}