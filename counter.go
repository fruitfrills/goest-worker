package goest_worker

import (
	"sync/atomic"
	"context"
	"time"
)


// Factory to create an atomic counter
var AtomicCounter CounterMiddleware = func () Counter {
	return &atomicCounter{}
}

type atomicCounter struct {
	size uint64
}

func (counter *atomicCounter) Increment() {
	atomic.AddUint64(&counter.size, 1)
}

func (counter *atomicCounter) Decrement() {
	atomic.AddUint64(&counter.size, ^uint64(0))
}

func (counter *atomicCounter) Len() uint64 {
	return atomic.LoadUint64(&counter.size)
}

// Waiting for counter will be equals to zero or context is closed
func (counter *atomicCounter) Wait(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Millisecond):
			if atomic.LoadUint64(&counter.size) == 0 {
				return
			}
		}
	}
	return
}