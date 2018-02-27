package main

import (
	worker "goest-worker" // "github.com/yobayob/goest-worker"
	"runtime"
	"fmt"
)

/*
Simple task with parameters and results
You can getting results by method `Results()` task.Run().Wait().Results()
 */
func sum(a, b int) (int) {
	fmt.Printf("%d + %d\n", a, b)
	return a+b
}

func main()  {
	pool := worker.New().Start(runtime.NumCPU())  	// create workers pool
	sumJob := pool.NewJob(sum)
	results, err := sumJob.Run(2, 256).Wait().Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("result is", results[0].(int))
	results, err = sumJob.Run(12, 12).Wait().Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("result is", results[0].(int))
	_, err = sumJob.Run("failed", 12).Wait().Result()
	if err != nil {
		fmt.Printf("catch error, %v\n", err)
	}
	pool.Stop()
}