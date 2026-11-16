package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Sender struct{
	urls [] string
}

func NewSender(sliceUrls []string) *Sender {
	return &Sender{
		urls: sliceUrls,
	}
}


type Result struct{
	Resp http.Response
	Err error
}
func (s *Sender) SendRequests(ctx context.Context) (<- chan Result){
	out := make(chan Result, 1)
	ctxWithCancel, cancel := context.WithCancel(ctx)
	attempts := make(chan Result, len(s.urls))
	for _, url := range s.urls {
		go func(targetURL string){
			response, err := httpGetRequest(ctxWithCancel, targetURL)
			if err != nil {
				attempts <- Result{Err: fmt.Errorf("%s: %w", targetURL, err)}
				return
			}
			body, err := io.ReadAll(response.Body)
			response.Body.Close()
			if  err != nil {
				attempts <- Result{Err: fmt.Errorf("%s: чтение тела ответа: %w", targetURL, err)}
				return
			}
			response.Body = io.NopCloser(bytes.NewReader(body))
			attempts <- Result{Resp: *response}
		}(url)
	}

	go func()  {
		defer close(out)
		defer cancel()
		var failures[] error
		for range s.urls{
			select{
			case <-ctx.Done():
				out <- Result{Err: ctx.Err()}
				return
			case attempt := <-attempts:
                if attempt.Err == nil {
                    out <- attempt
                    return
                }
                failures = append(failures, attempt.Err)
            }
		}
		out <- Result{
            Err: fmt.Errorf("все запросы завершились ошибкой: %w", errors.Join(failures...)),
        }
	}()
	return out
}

func httpGetRequest(ctx context.Context, rawUrl string) (*http.Response, error){
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawUrl, nil)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr){
			return nil, fmt.Errorf("Некорректный URL (%s): %w", urlErr.URL, err)
		}
		return nil, fmt.Errorf("Ошибка на этапе создания HTTP запроса: %w", err )
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		if errors.Is(err, context.Canceled){
			return nil, fmt.Errorf("Ошибка отмены запроса: %w", err)
		}
		if errors.Is(err, context.DeadlineExceeded){
			return nil, fmt.Errorf("Ошибка таймаута запроса: %w", err)
		}
		return nil, fmt.Errorf("Ошибка HTTP запроса: %w", err)
	}
	return response, nil
}

