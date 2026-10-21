package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func doAttempt(client *http.Client, method, url, key string) (int, string, error) {
	req, err := http.NewRequest(method, url, nil)
	if key != "" {
		req.Header.Set("idempotency-key", key)
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	return resp.StatusCode, resp.Header.Get("Retry-After"), nil
}

func isSuccess(status int) bool {
	return status >= 200 && status <= 399
}
func isRetryable(status int, err error) bool {
	if err != nil {
		return true
	}
	switch status {
	case 429, 500, 502, 503, 504:
		return true
	}
	return false
}

func methodAllowsRetry(method, key string) bool {
	if strings.ToUpper(method) == "POST" {
		return key != ""
	}
	return true
}

func pauseMs(attempt int, retryAfter string) int {
	if retryAfter != "" {
		sec, err := strconv.Atoi(retryAfter)
		if err == nil {
			return sec * 1000
		}
	}

	ceiling := 200
	for i := 1; i < attempt; i++ {
		ceiling *= 2
		if ceiling > 2000 {
			ceiling = 2000
			break
		}
	}
	return rand.Intn(ceiling + 1)
}

func main() {
	url := os.Args[1]

	fs := flag.NewFlagSet("retry", flag.ExitOnError)
	method := fs.String("method", "GET", "")
	maxAttempts := fs.Int("max-attempts", 5, "")
	key := fs.String("idempotency-key", "", "")
	fs.Parse(os.Args[2:])

	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: &http.Transport{DisableKeepAlives: true},
	}

	attempt := 1
	for {
		status, retryAfter, err := doAttempt(client, *method, url, *key)
		if err != nil {
			fmt.Println("attempt", attempt, "error", err)
		} else {
			fmt.Println("attempt", attempt, "status", status)
		}

		again := isRetryable(status, err) &&
			methodAllowsRetry(*method, *key) &&
			attempt < *maxAttempts
		if !again {
			result := "failure"
			exitCode := 1
			if err == nil && isSuccess(status) {
				result = "success"
				exitCode = 0
			}
			fmt.Println("result", result, "attempts", attempt)
			os.Exit(exitCode)
		}

		attempt++
		ms := pauseMs(attempt, retryAfter)
		fmt.Println("sleep_ms", ms)
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}

}
