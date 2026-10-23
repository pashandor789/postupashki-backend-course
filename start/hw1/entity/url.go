package entity

import (
	"fmt"
	"net/url"
)

type URL struct {
	url.URL
}

func NewURL(urlCmd string) (*URL, error){
	obj, err := url.Parse(urlCmd)
	if err != nil {
		return nil, fmt.Errorf("Invalid URL: %w", err)
	}
	return &URL {
		URL : *obj,
	}, nil
}