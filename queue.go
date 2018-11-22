package goest_worker

import "context"

// This is not a real queue with limited capacity and without support priority.
// If the queue is full, the process will wait before adding the task.
var ChannelQueue PoolQueue = func (ctx context.Context, capacity int) Queue {
	ctx, cancel := context.WithCancel(ctx)
	queue := &channelQueue{
		cancel: cancel,
		ctx: ctx,
		ch: make(chan jobCall, capacity),
	}
	return queue
}

type channelQueue struct {
	ctx context.Context
	cancel context.CancelFunc
	ch chan jobCall
}

func (c *channelQueue) SetCapacity(ctx context.Context, capacity int) {
	c.cancel()
	c.ctx, c.cancel = context.WithCancel(ctx)
	c.ch = make(chan jobCall, capacity)
}

func (c *channelQueue) Pop () (jobCall) {
	select {
	case j := <- c.ch:
		return j
	case <- c.ctx.Done():
		return nil
	}
}

func (c *channelQueue) Insert (job jobCall) () {
	select {
	case c.ch <- job:
		return
	case <- c.ctx.Done():
		return
	 }
}

func (c *channelQueue) Len () (uint64) {
	return uint64(len(c.ch))
}