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
func fiveSecondPassed(arg string) () {
	fmt.Println(arg, `passed`)
}
/*
Task without results, run on monday
 */
func everyMonday(name string) {
	fmt.Printf(`Hello, %s`, name)
}

func main()  {
	pool := worker.MainPool.Start(runtime.NumCPU())
	passTimeJob := worker.NewJob(fiveSecondPassed)
	passTimeJob.RunEvery(5 * time.Second, "5 seconds")					// run every 5 second
	passTimeJob.RunEvery(10 * time.Second, "10 seconds")
	passTimeJob.RunEvery("* * * * *", "One minunte")
	<- time.After(120 * time.Second)
	pool.Stop()										// stop pool
}