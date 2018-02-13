package main

import (
	worker "github.com/yobayob/goest-worker"
	"runtime"
	"fmt"
)

/*
Simple task with parameters and results
You can getting results by method `Results()` task.Run().Wait().Results()
 */
func sum(a, b int) (int) {
	fmt.Printf("%d + %d = %d \n", a, b, a+b)
	return a+b
}

func main()  {
	pool := worker.NewPool(runtime.NumCPU()).Start()  	// create workers pool
	task, err := worker.NewJob(sum, 2, 256)
	if err != nil {
		panic(err)
	}
	results := task.Run().Wait().Result()
	fmt.Println("result is", results[0].(int))
	pool.Stop()
}