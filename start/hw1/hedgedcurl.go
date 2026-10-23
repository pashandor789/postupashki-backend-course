package main

import (
	"context"
	"flag"
	"fmt"
	"hedgedcurl/service"
	"io"
	"os"
	"strings"
	"time"
)

func main() {
	var timeout int

	flag.IntVar(&timeout, "timeout", 15, "Таймаут запросов в секундах")
	flag.IntVar(&timeout, "t", 15, "Таймаут запросов в секундах (сокращенно)")
	flag.Parse()

	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	urls := flag.Args()
	sender := service.NewSender(urls)
	result := <-sender.SendRequests(ctx)
	if result.Err != nil {
		if result.Err == context.DeadlineExceeded{
			 fmt.Fprintln(os.Stderr, "Timeout Exceed")
        	os.Exit(228)
		}
		fmt.Fprintln(os.Stderr, result.Err)
    	os.Exit(1)
	}
	response := result.Resp
	body, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
    	fmt.Fprintf(os.Stderr, "Не удалось прочитать тело ответа: %v\n", err)
   		os.Exit(1)
	}
	var headers strings.Builder
	for key, values := range response.Header {
		for _, value := range values {
			fmt.Fprintf(&headers, "%s: %s\n", key, value)
		}
	}
	fmt.Printf("Полученный код ответа: %d заголовки: %s\n Тело: \n%s",response.StatusCode, headers.String(), body)
}
