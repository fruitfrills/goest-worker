# goest-worker
<img align="right" width="159px" src="https://raw.githubusercontent.com/fruitfrills/goest-worker/master/goest-worker.png">

Package goest-worker implements a priority queue of tasks and a pool of goroutines for launching and managing the states of these tasks.

## Features

- Lightweight queue of jobs with concurrent processing
- Periodic tasks (use cron expressions or time.Duration for repeating)
- Ability to add tasks with priority to the queue

## Installation

To install goest-worker, use

```sh
go get github.com/yobayob/goest-worker
```

## Getting started

```go
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
	results, err := sumJob.Run(1111, 2156).Wait().Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("result is", results[0].(int))
	divJob := pool.NewJob(div)
	results, err = divJob.Run( 144, 12).Wait().Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("result is", results[0].(int))
	pool.Stop()
}
```

### Examples
see examples https://github.com/yobayob/goest-worker/tree/master/examples


### Documentationd
godoc - https://godoc.org/github.com/fruitfrills/goest-worker
