package main

import (
	worker "goest-worker" // "github.com/yobayob/goest-worker"
	"fmt"
)

func sum(a, b int) (int) {
	return a + b
}

func mul(a, b int) (int) {
	return a * b
}

func sub(a, b int) (int) {
	return a - b
}

// bad example, but it works ;)

func main() {
	defaultPool := worker.MainPool.Start(1) // create workers pool
	task, err := worker.NewJob(sum, 2, 256)
	if err != nil {
		panic(err)
	}

	pool1 := worker.New().Start(2)
	pool2 := worker.New().Start(2)

	task1, err := pool1.NewJob(sub, 99, 9)

	if err != nil {
		panic(err)
	}
	task2, err := pool2.NewJob(mul, 25, 5)

	if err != nil {
		panic(err)
	}

	results, err := task.Run().Wait().Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("sum result is", results[0].(int))
	defaultPool.Stop()
	results, err = task1.Run().Wait().Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("sub result is", results[0].(int))
	results, err = task2.Run().Wait().Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("mul result is", results[0].(int))


	//restart
	defaultPool.Start(3)
	task3, err := defaultPool.NewJob(sum, 19, 5)
	results, err =  task3.Run().Wait().Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("sum result is", results[0].(int))
}
