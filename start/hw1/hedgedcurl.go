package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	defaultTimeoutSec = 15
	exitCodeTimeout   = 228
	exitCodeError    = 1
	exitCodeSuccess  = 0
	msgAllFailed     = "all requests failed"
)

func main() {
	var t int
	var h bool
	flag.IntVar(&t, "t", defaultTimeoutSec, "timeout in seconds")
	flag.IntVar(&t, "timeout", defaultTimeoutSec, "timeout in seconds")
	flag.BoolVar(&h, "h", false, "show help")
	flag.BoolVar(&h, "help", false, "show help")
	flag.Parse()

	if h {
		fmt.Fprintln(os.Stderr, "Usage: hedgedcurl [-t timeout] url1 [url2 ...]")
		flag.PrintDefaults()
		os.Exit(exitCodeSuccess)
	}

	urls := flag.Args()
	if len(urls) == 0 {
		fmt.Fprintln(os.Stderr, "hedgedcurl: provide at least one URL")
		os.Exit(exitCodeError)
	}

	first := make(chan *http.Response, 1)
	var mu sync.Mutex
	var failed int
	var hadTimeout bool
	timeout := time.Duration(t) * time.Second
	total := len(urls)
	client := &http.Client{Timeout: timeout}

	for _, url := range urls {
		go func(url string) {
			resp, err := client.Get(url)

			if err != nil {
				lastFailed := false
				mu.Lock()
				failed++
				if errors.Is(err, context.DeadlineExceeded) {
					hadTimeout = true
				} else if e, ok := err.(net.Error); ok && e.Timeout() {
					hadTimeout = true
				}
				if failed == total {
					lastFailed = true
				}
				mu.Unlock()
				if lastFailed {
					first <- nil
				}
				return
			}

			select {
				case first <- resp:
				default:
					resp.Body.Close()
			}
		}(url)
	}

	resp := <-first

	if resp == nil {
		fmt.Fprintln(os.Stderr, msgAllFailed)
		mu.Lock()
		exit228 := hadTimeout
		mu.Unlock()
		if exit228 {
			os.Exit(exitCodeTimeout)
		}
		os.Exit(exitCodeError)
	}
	defer resp.Body.Close()

	fmt.Fprintf(os.Stdout, "%s %s\r\n", resp.Proto, resp.Status)
	for k, v := range resp.Header {
		for _, vv := range v {
			fmt.Fprintf(os.Stdout, "%s: %s\r\n", k, vv)
		}
	}
	fmt.Fprint(os.Stdout, "\r\n")
	io.Copy(os.Stdout, resp.Body)
}
