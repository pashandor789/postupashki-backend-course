package main

import (
	"bufio"
	"log"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	response, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	if response != "OK\n" {
		log.Fatalf("unexpected response: %q", response)
	}
}
