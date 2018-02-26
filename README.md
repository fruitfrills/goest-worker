## goest-worker

Simple implementation a queue

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
	task, err := pool.NewJob(sum, 2, 256)
	if err != nil {
		panic(err)
	}
	err, results := task.Run().Wait().Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("result is", results[0].(int))
	pool.Stop()
}
```

### Features

- Lightweight queue of jobs with concurrent processing
- Periodical Jobs

### Examples
see examples https://github.com/yobayob/goest-worker/tree/master/examples 
