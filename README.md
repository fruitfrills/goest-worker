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
		fmt.Printf("catch error, %v\n", err)  // reflect: Call using string as type int
	}
	pool.Stop()
}
```

### Features

- Lightweight queue of jobs with concurrent processing
- Periodical Jobs

### Examples
see examples https://github.com/yobayob/goest-worker/tree/master/examples 
