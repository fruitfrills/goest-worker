package goest_worker

import "context"

type atomicCounter struct {
	ctx context.Context
	size uint64
}

func (counter *atomicCounter) Increment(i int) {

}

func (counter *atomicCounter) Decrement(i int) {

}

func (counter *atomicCounter) Len() {

}

func (counter *atomicCounter) Wait() {

}