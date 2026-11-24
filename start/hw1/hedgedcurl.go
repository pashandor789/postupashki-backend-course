package main

import (
	"fmt"
	"net/http"
)

const url = "https://www.google.com"

func main() {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	println("Response Status:", resp.Status)
}
