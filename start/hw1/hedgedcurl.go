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
	Status  string
	Proto   string
	Headers http.Header
	Body    []byte
	Err     error
}

func request(ctx context.Context, client *http.Client, url string, results chan<- Result) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		results <- Result{Err: fmt.Errorf("%s: %w", url, err)}
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		results <- Result{Err: fmt.Errorf("%s: %w", url, err)}
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		results <- Result{Err: fmt.Errorf("%s: %w", url, err)}
		return
	}

	results <- Result{
		Status:  resp.Status,
		Proto:   resp.Proto,
		Headers: resp.Header,
		Body:    body,
	}
}

func help() {
	fmt.Println("Использование: hedgedcurl [флаги] URL [URL ...]")
	fmt.Println("Отправляет GET-запросы одновременно и выводит первый готовый ответ.")
	fmt.Println("  -t, --timeout SECONDS  Общий таймаут в секундах (по умолчанию 15)")
	fmt.Println("  -h, --help             Показать справку")
}

func run() int {
	var timeout float64
	var showHelp bool
	flags := flag.NewFlagSet("hedgedcurl", flag.ContinueOnError)
	flags.Float64Var(&timeout, "t", 15, "таймаут в секундах")
	flags.Float64Var(&timeout, "timeout", 15, "таймаут в секундах")
	flags.BoolVar(&showHelp, "h", false, "справка")
	flags.BoolVar(&showHelp, "help", false, "справка")
	if err := flags.Parse(os.Args[1:]); err != nil {
		return 1
	}
	if showHelp {
		help()
		return 0
	}

	urls := flags.Args()
	if len(urls) == 0 {
		fmt.Fprintln(os.Stderr, "Укажите хотя бы один URL")
		return 1
	}
	duration := time.Duration(timeout * float64(time.Second))
	if timeout <= 0 || duration <= 0 {
		fmt.Fprintln(os.Stderr, "Таймаут должен быть положительным числом секунд")
		return 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	client := &http.Client{}
	defer client.CloseIdleConnections()

	results := make(chan Result, len(urls))
	for _, url := range urls {
		go request(ctx, client, url, results)
	}

	var lastErr error
	for range urls {
		select {
		case result := <-results:
			if result.Err != nil {
				lastErr = result.Err
				continue
			}

			cancel()
			fmt.Println(result.Proto, result.Status)
			for name, values := range result.Headers {
				for _, value := range values {
					fmt.Printf("%s: %s\n", name, value)
				}
			}
			fmt.Println()
			fmt.Print(string(result.Body))
			return 0
		case <-ctx.Done():
			fmt.Fprintln(os.Stderr, "Истекло время ожидания ответа")
			return 228
		}
	}

	if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(lastErr, context.DeadlineExceeded) {
		fmt.Fprintln(os.Stderr, "Истекло время ожидания ответа")
		return 228
	}
	fmt.Fprintln(os.Stderr, "Все запросы завершились ошибкой:", lastErr)
	return 1
}

func main() {
	os.Exit(run())
}
