// Package goest-worker implements a priority queue of tasks and a pool of goroutines for launching and managing the states of these tasks.
//
// Quick start
//
// Start pool with one or more workers
//
//     pool := worker.New().Start(context.TODO(), runtime.NumCPU())
//
// Create job for pool
//
//     sum := pool.NewJob(func (a, b int) (int) {
//         res := a / b
//         fmt.Printf("%d / %d\n", a, b)
//         return res
//     })
//
// Run current job
//
//     j := sum.Run(2, 2)
//
// Getting the result
//
//     var result int
//     j.Wait()
//     err := j.Result(&result)
//     fmt.PrintLn(result) // print 4
//
// Stop pool
//
//     pool.Stop()
//
// For more details, see the documentation for the types and methods.
//
package goest_worker