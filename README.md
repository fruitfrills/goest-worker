### GOest-worker

Simple implementation a queue


```
import (
  	worker "github.com/yobayob/goest_worker"
  	"time"
)

func main() {
    dispatcher := worker.NewPool(10)                                    // count of workers
    task, _ := worker.NewTask(func (arg string) {
        fmt.Println(arg)
    } (), "Hello, World!")                                              // create task
    dispatcher.AddPeriodicTask(time.Second * 5, *task)                   // add periodic task
    dispatcher.Start()                                                  // run
}
```