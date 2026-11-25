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

type result struct {
	body   []byte
	header http.Header
	status string
	err    error
}

func main() {
	var timeoutSeconds int
	flag.IntVar(&timeoutSeconds, "t", 15, "Timeout in seconds")
	flag.IntVar(&timeoutSeconds, "timeout", 15, "Timeout in seconds")
	var help bool
	flag.BoolVar(&help, "h", false, "Show help")
	flag.BoolVar(&help, "help", false, "Show help")
	flag.Parse()
	if help {
		fmt.Println("Usage: hedgedcurl [-t SECONDS | --timeout SECONDS] URL [URL ...]")
		fmt.Println("Default timeout: 15 seconds. Help: -h, --help.")
		return
	}

	urls := flag.Args()
	if timeoutSeconds <= 0 || len(urls) == 0 {
		os.Stderr.WriteString("Please provide at least one URL as a command-line argument.\n")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(timeoutSeconds)*time.Second,
	)
	defer cancel()
	results := make(chan result, len(urls))

	for _, url := range urls {
		go func(url string) {
			res := fetch(ctx, url)
			results <- res
		}(url)
	}

	for i := 0; i < len(urls); i++ {
		res := <-results
		if res.err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching URL: %v\n", res.err)
			continue
		}
		cancel()

		fmt.Printf("Status: %s\n", res.status)

		for name, values := range res.header {
			for _, value := range values {
				fmt.Printf("%s: %s\n", name, value)
			}
		}

		fmt.Println()
		if _, err := os.Stdout.Write(res.body); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing response body: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if ctx.Err() == context.DeadlineExceeded {
		fmt.Fprintln(os.Stderr, "Request timeout")
		os.Exit(228)
	}

	fmt.Fprintln(os.Stderr, "All requests failed")
	os.Exit(1)
}

func fetch(ctx context.Context, url string) result {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return result{err: err}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result{err: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result{err: err}
	}

	return result{
		body:   body,
		header: resp.Header,
		status: resp.Proto + " " + resp.Status,
	}
}
