package goest_worker

import (
	"testing"
	"fmt"
	"time"
	"runtime"
)


func TestWait(t *testing.T) {
	dispatcher := NewPool(3)                                    // count of workers
	dispatcher.Start()                                                  // run
	task, _ := NewTask(func (arg string) string {
		return arg + ", World!"
	}, "Hello")
	res := task.Run().Wait().Result()[0].(string)
	fmt.Println(res)
	if res != "Hello, World!" {
		t.Fail()
	}
	dispatcher.Stop()
}

func TestWait2(t *testing.T) {
	dispatcher := NewPool(3)                                    // count of workers
	dispatcher.Start()                                                  // run
	task, _ := NewTask(func (arg int) int {
		return 1 + arg
	}, 1)
	res := task.Run().Wait().Result()[0].(int)
	if res != 2 {
		t.Fail()
	}
	dispatcher.Stop()
}

func TestTicker(t *testing.T)  {
	dispatcher := NewPool(3)                                    // count of workers
	dispatcher.Start()
	task, _ := NewTask(func () {
		fmt.Println("I period!")
	}, )
	task.RunEvery(time.Second * 3)
	<- time.After(time.Second * 10)
	dispatcher.Stop()
}

func TestTicker2(t *testing.T)  {
	dispatcher := NewPool(3)                                    // count of workers
	dispatcher.Start()
	done := make(chan bool)
	task, _ := NewTask(func () {
		fmt.Println("I next!")
		done <- true
	}, )
	next := time.Now()
	next = next.Add(10 * time.Second)
	task.RunEvery(&Schedule{
		0,
		next.Hour(),
		next.Minute(),
		next.Second(),
		next.Nanosecond(),
	})
	<- done
	dispatcher.Stop()
}

func TestManyTasks(t *testing.T) {
	dispatcher := NewPool(runtime.NumCPU())                                    // count of workers
	dispatcher.Start()
	i := 1
	for i <=  1000 {
		task, _ := NewTask(func () {
			fmt.Println("Hello, task id #", i)
		}, )
		task.Run()
		i +=1
	}
	dispatcher.Stop()
}