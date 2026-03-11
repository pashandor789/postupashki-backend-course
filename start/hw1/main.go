package main

import (
	"context"
 	"errors"
 	"flag"
 	"fmt"
 	"io"
 	"net/http"
 	"os"
 	"time"
)

type Result struct {
	resp *http.Response
 	err  error
}

func fetch(ctx context.Context, url string, ch chan<- Result) {
 	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
 	if err != nil {
  		ch <- Result{nil, err}
  		return
 	}

 	client := &http.Client{}
 	resp, err := client.Do(req)

 	ch <- Result{resp, err}
}

func main() {
 	timeout := flag.Int("t", 15, "timeout in seconds")
 	help := flag.Bool("h", false, "help")
 	flag.BoolVar(help, "help", false, "help")

 	flag.Parse()

 	if *help {
  		fmt.Println("Usage: hedgedcurl [-t seconds] url1 url2 ...")
  		return
 	}

 	urls := flag.Args()
 	if len(urls) == 0 {
  		fmt.Println("no urls provided")
  		os.Exit(1)
 	}

 	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout)*time.Second)
 	defer cancel()

 	resultChan := make(chan Result)

 	for _, url := range urls {
  		go fetch(ctx, url, resultChan)
	}

 	errorsCount := 0

 	for {
  		select {
  		case res := <-resultChan:

   			if res.err != nil {
    			errorsCount++
    			if errorsCount == len(urls) {
     				fmt.Println("all requests failed")
     				return
    			}
    			continue
   			}

   			cancel()

   			defer res.resp.Body.Close()

  			body, _ := io.ReadAll(res.resp.Body)

   			fmt.Println(res.resp.Status)

   			for k, v := range res.resp.Header {
    			for _, val := range v {
     				fmt.Printf("%s: %s\n", k, val)
    			}	
  			}

   			fmt.Println()
   			fmt.Print(string(body))

   			return

  		case <-ctx.Done():
   			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
    			os.Exit(228)
   			}
   			return
  		}
 	}
}