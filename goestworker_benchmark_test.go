package goest_worker

import (
	"testing"
	"time"
	"context"
	"sync"
)

var testsCountOfTasks = 10000

func demoFunc() error {
	time.Sleep( 10 * time.Millisecond)
	return nil
}

// test clean goroutine
func BenchmarkGoroutine(b *testing.B)  {
	var wg sync.WaitGroup
	wg.Add(testsCountOfTasks)
	b.StartTimer()
	for j := 0; j < testsCountOfTasks; j++ {
		go func() {
			demoFunc()
			wg.Done()
		}()
	}
	wg.Wait()
	b.StopTimer()
}

func BenchmarkPoolGoroutine10(b *testing.B)  {
	var wg sync.WaitGroup

	ch := make(chan struct{}, 10)
	wg.Add(testsCountOfTasks)

	b.StartTimer()

	for i := 0; i < 10; i++ {
		go func() {
			for {
				<- ch
				demoFunc()
				wg.Done()
			}
		}()
	}

	for j := 0; j < testsCountOfTasks; j++ {
		ch <- struct{}{}
	}
	wg.Wait()
	b.StopTimer()
}

// test default goroutine pool
func BenchmarkPoolGoroutine1000(b *testing.B)  {
	var wg sync.WaitGroup

	ch := make(chan struct{}, 1000)
	wg.Add(testsCountOfTasks)

	b.StartTimer()

	for i := 0; i < 1000; i++ {
		go func() {
			for {
				<- ch
				demoFunc()
				wg.Done()
			}
		}()
	}

	for j := 0; j < testsCountOfTasks; j++ {
		ch <- struct{}{}
	}
	wg.Wait()
	b.StopTimer()
}

func BenchmarkPool10(b *testing.B) {
	pool := New().Start(context.TODO(), 10)
	defer pool.Stop()
	job := pool.NewJob(demoFunc)
	b.StartTimer()
	for i := 0; i < 10000; i ++ {
		job.Run()
	}
	pool.Wait()
	b.StopTimer()
}

func BenchmarkPool1000(b *testing.B) {
	pool := New().Start(context.TODO(), 1000)
	defer pool.Stop()
	job := pool.NewJob(demoFunc)
	b.StartTimer()
	for i := 0; i < testsCountOfTasks; i ++ {
		job.Run()
	}
	pool.Wait()
	b.StopTimer()
}