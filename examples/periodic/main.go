package main

import (
	worker "goest-worker" // "github.com/yobayob/goest-worker"
	"runtime"
	"time"
	"fmt"
)

/*
Simple task with parameters and results
You can getting results by method `Results()` task.Run().Wait().Results()
 */
func simplePeriodicTask(a, b int) (int) {
	fmt.Printf(`%d * %d = %d`, a, b, a*b)
	return a*b
}

/*
Task without results, run on monday
 */
func everyMonday(name string) {
	fmt.Printf(`Hello, %s`, name)
}

func main()  {
	pool := worker.MainPool.Start(runtime.NumCPU())
	task, err := worker.NewJob(simplePeriodicTask, 2, 256)
	if err != nil {
		panic(err)
	}
	task.RunEvery(5 * time.Second)					// run every 5 second

	task, err = worker.NewJob(everyMonday, "Evgenyi")
	if err != nil {
		panic(err)
	}
	task.RunEvery("@hourly")						// run monday at 12:30
	<- time.After(30 * time.Second)
	pool.Stop()										// stop pool
}