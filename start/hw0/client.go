package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func check_server() error {
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		return fmt.Errorf("не удалось подключиться к серверу: %w", err)
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	response, _ := reader.ReadString('\n')

	if response != "OK\n" {
		return fmt.Errorf("Неправильный ответ %q", response)
	} else {
		return nil
	}

}

func main() {
	if err := check_server(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
