# goest-worker
<img align="right" width="159px" src="https://raw.githubusercontent.com/fruitfrills/goest-worker/master/goest-worker.png">

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

func diff(a, b int) (int) {
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
	diffJob := pool.NewJob(diff)
	results, err = diffJob.Run( 144, 12).Wait().Result()
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