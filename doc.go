// Package goest-worker implements a priority queue of tasks and a pool of goroutines for launching and managing the states of these tasks.
//
// Quick start
//
//      pool := worker.New().Start(context.TODO(), runtime.NumCPU())
//
// This will start pool with one or more workers
//
//     sum := pool.NewJob(func (a, b int) (int) {
//         res := a / b
//         fmt.Printf("%d / %d\n", a, b)
//         return res
//     })
//
// This will create job for pool
//
//     j := sum.Run(2, 2)
//
// This will run current job
//
//     j.Wait()
//     results, err := j.Result()
//     fmt.PrintLn(results[0].(int)) // print 4
//
// So you can get the result
//
//     pool.Stop()
//
// This will stop pool
//
// For more details, see the documentation for the types and methods.

package goest_worker