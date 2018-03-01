package local

import (
	"testing"
	"runtime"
)

func BenchmarkPool(b *testing.B) {
	b.SetBytes(2)
	ch := make(chan int, b.N)
	done := make(chan struct{})
	numCPUs := runtime.NumCPU()

	go func() {
		for i := 0; i < b.N; i++ {
			<-ch
		}
		done <- struct{}{}
		close(ch)
		close(done)
	}()
	pool := New().Start(numCPUs)
	defer pool.Stop()
	job := pool.NewJob(func(i int) {
		ch <- i
	})
	for i := 0; i < b.N; i++ {
		job.Run(i)
	}
	<-done
}