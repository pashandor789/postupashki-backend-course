package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	var timeout int

	flag.IntVar(&timeout, "t", 5, "This function makes a timeout to your url(s)")
	flag.IntVar(&timeout, "timeout", 5, "This function makes a timeout to your url(s)")

	flag.Parse()

	urls := flag.Args()

	if len(urls) == 0{
		fmt.Println("You must provide any url")
		os.Exit(1)
	}

	successCh := make(chan string)
	errorCh := make(chan error)

	for _, url := range urls{
		go HttpGet(url, successCh, errorCh)
	}

	errorCount := 0

	for {
		select {
		case response := <-successCh:
			fmt.Println(response)
			os.Exit(0)
		case <-errorCh:
			errorCount++
			if errorCount == len(urls) {
				fmt.Println("All urls were not read")
				os.Exit(1)
			}
		case <-time.After(time.Duration(timeout) * time.Second):
			fmt.Println("Ran out of time")
			os.Exit(228)
		}
	}
}

func HttpGet(url string, successCh chan string, errorCh chan error) {
	resp, err := http.Get(url)

	if err != nil{
		errorCh <- err
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400{
		errorCh <- errors.New(fmt.Sprintf("Url was not read with status: %v", resp.Status))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil{
		errorCh <- err
		return
	}

	res := fmt.Sprintf("Your url request went with code: %v (%v) with body: %v", resp.StatusCode, resp.Status, string(body))

	successCh <- res
}
