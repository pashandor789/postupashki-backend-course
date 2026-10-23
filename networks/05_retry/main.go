package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

var retryCodes = map[int]bool{429: true, 500: true, 502: true, 503: true, 504: true}

func main() {
	url := os.Args[1]
	method := "GET"
	maxAttempts := 5
	key := ""
	for i := 2; i+1 < len(os.Args); i += 2 {
		switch os.Args[i] {
		case "--method":
			method = os.Args[i+1]
		case "--max-attempts":
			maxAttempts, _ = strconv.Atoi(os.Args[i+1])
		case "--idempotency-key":
			key = os.Args[i+1]
		}
	}

	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: &http.Transport{DisableKeepAlives: true},
	}
	canRetry := method != "POST" || key != ""

	for attempt := 1; ; attempt++ {
		req, _ := http.NewRequest(method, url, nil)
		if key != "" {
			req.Header.Set("Idempotency-Key", key)
		}

		retry := false
		retryAfter := -1
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("attempt", attempt, "error", err.Error())
			retry = true
		} else {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			fmt.Println("attempt", attempt, "status", resp.StatusCode)
			if resp.StatusCode >= 200 && resp.StatusCode <= 399 {
				fmt.Println("result success attempts", attempt)
				return
			}
			retry = retryCodes[resp.StatusCode]
			if seconds, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil {
				retryAfter = seconds
			}
		}

		if !retry || !canRetry || attempt >= maxAttempts {
			fmt.Println("result failure attempts", attempt)
			os.Exit(1)
		}

		sleepMs := retryAfter * 1000
		if retryAfter < 0 {
			sleepMs = rand.Intn(min(2000, 200<<attempt) + 1)
		}
		fmt.Println("sleep_ms", sleepMs)
		time.Sleep(time.Duration(sleepMs) * time.Millisecond)
	}
}
