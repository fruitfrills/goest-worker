package main

import (
	"github.com/yobayob/goest-worker"
	"os"
	"net/http"
	"io"
	"context"
	"bufio"
	"strings"
	"fmt"
	"path/filepath"
	"strconv"
	"runtime"
)

func DownloadFile(uri, fp string) error {
	fmt.Println(uri, "process...")
	resp, err := http.Get("http://" + uri)
	if err != nil {
		fmt.Println(uri, "return error", err.Error())
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(fp)
	if err != nil {
		fmt.Println(uri, "return error", err.Error())
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(uri, "success downloaded")
	return nil
}

func getUriAndPriority(s string) (uri string, priority int) {
	args := strings.Split(strings.TrimSpace(s), " ")
	uri = args[0]
	priority, _ = strconv.Atoi(args[1])
	return uri, priority
}

func main()  {

	// create new pool
	pool := goest_worker.New().Use(
		goest_worker.PriorityQueue, // use priority queue for adding all task (this is not necessary, priority queue is defaul queue)
		goest_worker.AtomicCounter, // use atomic counter for pool.Wait()
	).Start(context.TODO(), runtime.NumCPU())

	// Stop the pool when all tasks are completed.
	defer pool.Stop()

	downloadJob := pool.NewJob(DownloadFile)
	file, err := os.Open("./top.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		uri, priority := getUriAndPriority(scanner.Text())
		fmt.Printf("put new job to pool %s with priority %d\r\n", uri, priority)
		downloadJob.RunWithPriority(priority, uri, filepath.Join("/tmp/", uri))
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println("Waiting for the pool to complete all tasks.")
	pool.Wait()
}