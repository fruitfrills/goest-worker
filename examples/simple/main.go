package main

import (
	worker "github.com/yobayob/goest-worker"
	"runtime"
	"fmt"
	"context"
)
/*
Simple task with parameters and results
You can getting results by method `Results()` task.Run().Wait().Results()
 */
func sum(a, b int) (int) {
	res := a+b
	fmt.Printf("%d + %d = %d\n", a, b, res)
	return res
}

func div(a, b int) (int) {
	res := a / b
	fmt.Printf("%d / %d\n", a, b)
	return res
}


func main()  {
	pool := worker.New().Start(context.TODO(), runtime.NumCPU())  	// create workers pool
	sumJob := pool.NewJob(sum)
	var result int
	err := sumJob.Run(1111, 2156).Wait().Results(&result)
	if err != nil {
		panic(err)
	}
	fmt.Println("result is", result)
	divJob := pool.NewJob(div)
	err = divJob.Run( 144, 12).Wait().Results(&result)
	if err != nil {
		panic(err)
	}
	fmt.Println("result is", result)
	pool.Stop()
}
