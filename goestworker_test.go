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
	var result int
	err := j.Results(&result)
	if err != nil {
		t.Error(err)
		return
	}

	if result != 30 {
		t.Error("incorect results")
		return
	}
}

func Test_processor(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	q :=  ChannelQueue(ctx, 0)
	ch := make(workerPoolType)
	go processor(ctx, q, ch)

	q.Insert(&mockJobCall{})
	if q.Len() != 0 {
		t.Error("failed processor")
		return
	}
}

func TestPool_Resize(t *testing.T) {
	pool := &pool{}
	pool.Start(context.TODO(), 2)
	pool.Resize(8)
	if len(pool.workers) != 8 {
		t.Error("pool worker count is not 10; pool is not increasing")
		return
	}
}

func TestPool_Resize2(t *testing.T) {
	pool := &pool{}
	pool.Start(context.TODO(), 8)
	pool.Resize(2)
	if len(pool.workers) != 2 {
		t.Error("pool worker count is not 2; pool is not reducing")
		return
	}
}

func TestPool_Resize3(t *testing.T) {
	pool := &pool{}
	pool.Start(context.TODO(), 2)
	pool.Resize(2)
	if len(pool.workers) != 2 {
		t.Error("pool worker count is not 2; pool is changed")
		return
	}
}
