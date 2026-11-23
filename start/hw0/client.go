package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	c, err := net.Dial("tcp", "localhost:8080")

	if err != nil {
		log.Fatalf("Ошибка соединения: %v", err)
	}

	defer c.Close()

	b, err := io.ReadAll(c)
	if err != nil {
		log.Printf("Ошибка чтения: %v", err)
		return
	}

	response := string(b)

	if response != "OK\n" {
		log.Fatalf("Вернулся не ОК((\n")
	}
	fmt.Print(response)
}
