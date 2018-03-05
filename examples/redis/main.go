package main

import (
	"runtime"
	"fmt"
	"goest-worker"
	"goest-worker/common"
	"goest-worker/redis"
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
	sumJob := pool.NewJob(sum)
	results, err := sumJob.Bind(true).SetMaxRetry(2).Run(1111, 2156).Wait().Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("result is", results[0].(uint))
	diffJob := pool.NewJob(diff)
	results, err = diffJob.Run( 144, 12).Wait().Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("result is", results[0].(int16))
	pool.Stop()
}