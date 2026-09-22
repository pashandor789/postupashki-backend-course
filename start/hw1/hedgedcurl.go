package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type fetchResult struct {
	url     string
	body    []byte
	status  string
	headers http.Header
	err     error
}

func main() {

	var timeout int
	flag.IntVar(&timeout, "t", 15, "timeout in seconds")
	flag.IntVar(&timeout, "timeout", 15, "timeout in seconds (alias for -t)")
	// Help тоже добавим явно:
	var help bool
	flag.BoolVar(&help, "h", false, "show help")
	flag.BoolVar(&help, "help", false, "show help")
	flag.Parse()

	if help {
		flag.Usage() // This prints the default usage text (which includes your flags)
		return       // Exit the program
	}

	urls := flag.Args()

	if len(urls) == 0 {
		fmt.Println("No Urls was entered!!!!")
		flag.Usage()
		os.Exit(1)
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)

	defer cancelFunc()

	ch := make(chan fetchResult, len(urls))

	for _, url := range urls {
		go func(url string) {
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

			if err != nil {
				ch <- fetchResult{url: url, err: err}
			} else {
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					ch <- fetchResult{url: url, err: err}
					return
				}
				body, err := io.ReadAll(resp.Body)
				defer resp.Body.Close()

				ch <- fetchResult{url: url, body: body, status: resp.Status, headers: resp.Header, err: err}
			}

		}(url)
	}

	errorCount := 0
	for {
		select {
		case res := <-ch:
			if res.err != nil {
				errorCount++
				if errorCount == len(urls) {
					fmt.Println("All requests failed")
					os.Exit(1)
				}
			} else {
				fmt.Println(res.status) // e.g., "200 OK"
				for key, values := range res.headers {
					for _, value := range values {
						fmt.Printf("%s: %s\n", key, value)
					}
				}
				fmt.Println()               // Blank line between headers and body
				fmt.Print(string(res.body)) // Print the raw body
				os.Exit(0)
			}

		case <-ctx.Done():
			fmt.Println("Request Timed out!!!!")
			os.Exit(228)
		}
	}
}
