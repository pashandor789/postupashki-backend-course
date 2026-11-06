package main

import (
	"io"
	"net"
	"os"
)

func main() {
	resp, err := net.Dial("tcp", "localhost:8080")

	if err != nil {
		os.Exit(-1)
	}

	defer resp.Close()

	body, err := io.ReadAll(resp)

	if err != nil {
		os.Exit(-1)
	}

	if string(body) != "OK\n" {
		os.Exit(-1)
	}
}
