package goest_worker

import (
	"testing"
	"context"
)


func TestNew(t *testing.T) {
	pool := New()
	pool.Start(context.TODO(), 8)
	job := pool.NewJob(func (a, b int) int {
		return a * b
	})
	j := job.Run(5,6)
	<- j.Context().Done()
	results, err := j.Result()
	if err != nil {
		t.Error(err)
	}

	if len(results) != 1 {
		t.Error("results lentgh incorrect")
	}

	if results[0].(int) != 30 {
		t.Error("incorect results")
	}
}

func TestPool_Resize(t *testing.T) {
	pool := &pool{}
	pool.Start(context.TODO(), 2)
	pool.Resize(8)
	if len(pool.workers) != 8 {
		t.Error("pool worker count is not 10; pool is not increasing")
	}
}

func TestPool_Resize2(t *testing.T) {
	pool := &pool{}
	pool.Start(context.TODO(), 8)
	pool.Resize(2)
	if len(pool.workers) != 2 {
		t.Error("pool worker count is not 2; pool is not reducing")
	}
}

func TestPool_Resize3(t *testing.T) {
	pool := &pool{}
	pool.Start(context.TODO(), 2)
	pool.Resize(2)
	if len(pool.workers) != 2 {
		t.Error("pool worker count is not 2; pool is changed")
	}
}
