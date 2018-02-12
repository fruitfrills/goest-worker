### Goest worker

Simple implementation a queue


```
import (
  	worker "github.com/yobayob/goest_worker"
  	"time"
)

func doSomething(a, b int) (int) {
    return a + b
}

func main() {
    pool := worker.NewPool(10).Start()                                  // count of workers
    task, _ := worker.NewTask(doSomething, 1, 2)                        // create task
    results := task.Run().Wait().Results()                              // run, wait, result
    summ := results[0].(int)
    print(summ)                                                         // 3
}
```

see examples 