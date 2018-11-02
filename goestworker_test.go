package goest_worker

import (
	"testing"
	"context"
	"time"
	"fmt"
)

func Multiplie(a, b int) int {
	return a * b
}

func TestNew(t *testing.T) {
	pool := New()
	pool.Start(context.TODO(), 8)
	job := pool.NewJob(Multiplie)
	j := job.Run(5,6)
	<- j.Context().Done()
	fmt.Println(j.Result())
	time.Sleep(time.Second)
}