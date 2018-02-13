## goest-worker

Simple implementation a queue

### Getting started

```
go get github.com/yobayob/goest-worker
```

### Basic usage

```
package main

import (
	worker "github.com/yobayob/goest-worker"
	"runtime"
	"fmt"
)

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
```

### Features

- Lightweight queue of jobs with concurrent processing
- Periodical Jobs

### Examples
see examples https://github.com/yobayob/goest-worker/tree/master/examples 
