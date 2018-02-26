package main

import (
	worker "goest-worker" // "github.com/yobayob/goest-worker"
	"runtime"
	"fmt"
	"time"
)

/*
Simple task with parameters and results
You can getting results by method `Results()` task.Run().Wait().Results()
 */
func sum(a, b int) (int) {
	return a+b
}

func main()  {
	pool := worker.New().Start(runtime.NumCPU())  	// create workers pool
	task, err := pool.NewJob(sum, 2, 256)
	if err != nil {
		panic(err)
	}
	results, err := task.Run().Wait().Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("result is", results[0].(int))
	pool.Stop()
}