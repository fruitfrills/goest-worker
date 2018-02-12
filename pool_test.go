package goest_worker

import (
	//"time"
	"testing"
)
//
//func TestExample(t *testing.T) {
//	done := make(chan bool)
//	dispatcher := NewPool(10)                                    // count of workers
//	task, _ := NewTask(func (arg string) {
//		fmt.Println(arg)
//		done <- true
//	}, "Hello, World!")                                    	// create task
//	dispatcher.AddPeriodicTask(time.Second * 5, *task)                   // add periodic task
//	dispatcher.Start()                                                  // run
//	select {
//		case <- time.After(time.Second * 10):
//			t.Fail()
//		case <- done:
//			dispatcher.Stop()
//	}
//}


func TestWait(t *testing.T) {
	dispatcher := NewPool(3)                                    // count of workers
	dispatcher.Start()                                                  // run
	task, _ := NewTask(func (arg string) string {
		return arg + ", World!"
	}, "Hello")
	res := task.Do().Wait().Result()[0].(string)
	if res != "Hello, World!" {
		t.Fail()
	}
	dispatcher.Stop()
}