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


func main()  {
	pool := goest_worker.New().Use(goest_worker.ChannelQueue).Start(context.TODO(), 8)
	defer pool.Stop()

	downloadJob := pool.NewJob(DownloadFile)
	file, err := os.Open("./top.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		args := strings.Split(strings.TrimSpace(scanner.Text()), " ")
		fmt.Println("put new job to pool", args[0])
		priority, _ := strconv.Atoi(args[1])
		uri := args[0]
		downloadJob.RunWithPriority(priority, uri, filepath.Join("/tmp/", uri))
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println("Waiting for the pool to complete all tasks.")
	pool.Wait()
}