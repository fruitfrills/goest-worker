package main

import (
	"runtime"
	"fmt"
	"github.com/yobayob/goest-worker"
	"github.com/yobayob/goest-worker/common"
	"github.com/yobayob/goest-worker/redis"
	"time"
)
/*
Simple task with parameters and results
You can getting results by method `Results()` task.Run().Wait().Results()
 */
func sum(job common.JobInjection, a, b uint) (uint) {
	res := a+b
	fmt.Printf("%d + %d = %d\n", a, b, res)
	return res
}

func diff(a, b int16) (int16) {
	res := a / b
	fmt.Printf("%d / %d\n", a, b)
	return res
}

func main()  {
	pool := goest_worker.New(&redis.RedisBackend{
		Addr: "localhost:6379",
		DbNum: 6,
		QueueName: "tasks",
		Password: "",
	}).Start(runtime.NumCPU())  	// create workers pool
	sumJob := pool.NewJob(sum).Bind(true)
	i := 0
	go func() {
		for i < 1000 {
			sumJob.Run(i, i).Wait()
			i += 1
		}
	}()

	diffJob := pool.NewJob(diff)
	results, err := diffJob.Run( 144, 12).Wait().Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("result is", results[0].(int16))
	time.Sleep(time.Second * 5)
	pool.Stop()
}