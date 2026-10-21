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
	Proto  string
	Status string
	Header http.Header
	Body   []byte
	Err    error
}

type Config struct {
	Timeout int
	URLs    []string
}

func main() {
	conf, err := parseConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.Timeout)*time.Second)
	defer cancel()
	res, err := runHedged(ctx, conf.URLs)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		if errors.Is(err, context.DeadlineExceeded) {
			os.Exit(228)
		}

		os.Exit(1)
	}
	err = printResponse(res)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// 3. Обработка аргументов
func parseConfig() (Config, error) {
	var timeout int
	flag.IntVar(&timeout, "timeout", 15, "timeout in seconds")
	flag.IntVar(&timeout, "t", 15, "timeout in seconds")

	flag.Parse()
	if len(flag.Args()) == 0 {
		return Config{}, errors.New("no URLs specified")
	}
	if timeout <= 0 {
		return Config{}, errors.New("timeout must be greater than zero")
	}
	return Config{Timeout: timeout, URLs: flag.Args()}, nil
}

// 4. Управление параллельными запросами
func runHedged(ctx context.Context, urls []string) (Result, error) {
	results := make(chan Result, len(urls))

	for _, url := range urls {
		go func(u string) {
			res := fetch(ctx, u)
			results <- res
		}(url)
	}
	for i := 0; i < len(urls); i++ {
		select {
		case resp := <-results:
			if resp.Err != nil {
				continue
			}
			return resp, nil
		case <-ctx.Done():
			return Result{}, ctx.Err()
		}
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	return Result{}, errors.New("all requests failed")
}

// 5. Выполнение одного HTTP-запроса
func fetch(ctx context.Context, url string) Result {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Result{Err: err}
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Result{Err: err}
	}
	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{Err: err}
	}
	return Result{resp.Proto, resp.Status, resp.Header, bytes, err}
}

// 6. Вывод HTTP-ответа
func printResponse(result Result) error {
	fmt.Println(result.Proto, result.Status)
	for key, values := range result.Header {
		for _, value := range values {
			fmt.Printf("%s: %s\n", key, value)
		}
	}

	_, err := os.Stdout.Write(result.Body)
	if err != nil {
		return err
	}
	return nil
}
