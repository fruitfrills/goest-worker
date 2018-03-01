package main

import (
	worker "goest-worker/local" // "github.com/yobayob/goest-worker"
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
	pool := worker.MainPool
	passTimeJob := worker.NewJob(pass)
	passTimeJob.RunEvery(5 * time.Second, "5 seconds")					// run every 5 second
	passTimeJob.RunEvery(10 * time.Second, "10 seconds")
	passTimeJob.RunEvery("* * * * *", "One minunte")

	worker.NewJob(failFunc).RunEvery(3 * time.Second)

	pool.Start(5)
	<- time.After(120 * time.Second)
	pool.Stop()										// stop pool
}