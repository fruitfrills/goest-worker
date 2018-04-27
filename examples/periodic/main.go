package main

import (
	worker "github.com/yobayob/goest-worker" // "github.com/yobayob/goest-worker"
	"github.com/yobayob/goest-worker/redis"
	"time"
	"fmt"
)

/*
Simple task with parameters and results
You can getting results by method `Results()` task.Run().Wait().Results()
 */
func failFunc()  {
	<-time.After(9 * time.Second)
	fmt.Println("I'm func!")
}


func pass(arg string) () {
	fmt.Println(arg, `passed`)
}

func main()  {
	pool := worker.New(&redis.RedisBackend{
		Addr: "localhost:6379",
		DbNum: 6,
		QueueName: "tasks",
		Password: "",
	})
	passTimeJob := pool.NewJob(pass)
	passTimeJob.RunEvery(5 * time.Second, "5 seconds")					// run every 5 second
	passTimeJob.RunEvery(10 * time.Second, "10 seconds")
	passTimeJob.RunEvery("* * * * *", "One minunte")

	pool.NewJob(failFunc).RunEvery(3 * time.Second)

	pool.Start(5)
	<- time.After(120 * time.Second)
	pool.Stop()										// stop pool
}