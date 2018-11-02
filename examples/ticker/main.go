package main

import (
	"fmt"
	"time"
	"github.com/yobayob/goest-worker"
	"context"
)

func task(interval string)  {
	fmt.Println(interval)
}

func main() {
	pool := goest_worker.New()

	pJob := pool.NewJob(task)
	pJob.RunEvery(time.Second * 10, "10s")
	pJob.RunEvery(time.Second * 20, "20s")
	pJob.RunEvery(time.Second * 30, "30s")
	pool.Start(context.TODO(), 5)
	defer pool.Stop()
	<- time.After(time.Minute)
}