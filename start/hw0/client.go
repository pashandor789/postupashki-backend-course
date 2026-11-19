package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal("Error connecting to server:", err)
	}

	fmt.Println("Connected to server:", conn.RemoteAddr().String())

	answer, err := io.ReadAll(conn)
	conn.Close()
	if err != nil {
		log.Fatal("Error reading from server:", err)
	}

	if string(answer) != "OK\n" {
		log.Fatal("Unexpected response from server:", string(answer))
	}
	fmt.Printf("Received from server: %q", answer)
}
