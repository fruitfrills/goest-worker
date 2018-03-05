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
	"strconv"
)

var poolMap sync.Map

// redis backend
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
	jobs		sync.Map
}


func (backend *RedisBackend) AddJobToPool(p common.PoolInterface,job common.JobInstance) () {

	redisJob := job.(*jobRedisFuncInstance)
	redisJob.PubName = redisJob.job.name + strconv.Itoa(int(time.Now().Unix()))
	redisJob.pubSubChannel = backend.conn.Subscribe(redisJob.PubName)
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

			if err = val.Err(); err != nil {
				if err == db.Nil {
					continue
				} else {
					log.Fatal(`redis error`)
				}
			}
			var job common.JobInstance
			if job, err = backend.unmarshallJob(val.Val()[1]); err != nil {
				log.Println(`Redis unmarshall error`, err)
				continue
			}
			// get free worker and send task
			worker := <-backend.workerPool
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
	poolMap.Delete(backend.QueueName)
	return p
}

func (backend *RedisBackend) Start(p common.PoolInterface, count int) (common.PoolInterface) {
	// if dispatcher started - do nothing
	if !p.IsStopped() {
		return p
	}

	if _, ok := poolMap.Load(backend.QueueName); ok {
		panic(`queue exists`)
	}

	// create connect
	backend.conn = db.NewClient(&db.Options{
		Addr:     backend.Addr,
		Password: backend.Password,
		DB:       backend.DbNum,
	})

	backend.workerPool = make(common.WorkerPoolType, count)
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
	poolMap.Store(backend.QueueName, backend)
	return p
}

func (backend *RedisBackend) NewJob(p common.PoolInterface, taskFn interface{}) (job common.Job) {

	fn := reflect.ValueOf(taskFn)
	fnType := fn.Type()

	if fnType.Kind() != reflect.Func {
		panic("job is not func")
	}

	name := runtime.FuncForPC(fn.Pointer()).Name()
	job = &jobRedisFunc{
		name: name,
		fn: fn,
		maxRetry: -1,
		pool: p,
		PoolName: backend.QueueName,
	}
	backend.jobs.Store(name, job)
	return job
}

func (backend *RedisBackend) Register(p common.PoolInterface, name string, taskFn interface{}) (job common.Job) {
	if !p.IsStopped() {
		panic(`pool is started`)
	}
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
		PoolName: backend.QueueName,
	}
	backend.jobs.Store(name, job)
	return job
}


// parse job
func (backend *RedisBackend) unmarshallJob(value string) (common.JobInstance, error){
	var jobInstance jobRedisFuncInstance
	err := json.Unmarshal([]byte(value), &jobInstance)
	if val, ok := backend.jobs.Load(jobInstance.JobName); ok {
		jobInstance.job = val.(*jobRedisFunc)
	}
	if val, ok := poolMap.Load(jobInstance.PoolName); ok {
		jobInstance.conn = val.(*RedisBackend).conn
	}
	return &jobInstance, err
}