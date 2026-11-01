package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func runClient() error {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}

	defer conn.Close()
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("read error: %w", err)
	}
	if response != "OK\n" {
		return fmt.Errorf("invalid server response: %q", response)
	}
	return nil
}

func main() {
	if err := runClient(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
