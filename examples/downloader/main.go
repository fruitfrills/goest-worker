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
)

func DownloadFile(self goest_worker.JobInstance, uri string) error {


	// this is for example
	select {
	case <-self.Context().Done():
		 return nil
	default:
		break
	}

	fmt.Println("start download", uri)
	// Get the data
	resp, err := http.Get("http://" + uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create("/dev/null")
	if err != nil {
		fmt.Println(uri, "return error", err.Error())
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println(uri, "return error", err.Error())
		return err
	}
	fmt.Println("download success", uri)
	return nil
}

func main()  {
	pool := goest_worker.New()
	defer pool.Stop()
	downloadJob := pool.NewJob(DownloadFile).Bind(true)
	pool.Start(context.TODO(), 8)
	file, err := os.Open("./top.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	i := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if i > 50 {
			break
		}
		s := strings.TrimSpace(scanner.Text())
		downloadJob.Run(s)
		fmt.Println("put to pool:", s)
		i += 1

	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	pool.Wait()
}