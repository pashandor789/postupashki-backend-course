package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func main() {
	const server_address = "localhost:8080"

	conn, err := net.Dial("tcp", server_address)
	if err != nil {
		log.Fatalf("Can't connect to server: %v", err)
	}
	defer conn.Close()

	fmt.Println("Successfully connected to", server_address)

	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Can't get answer from server: %v", err)
	}

	fmt.Printf("Got answer from server: %q\n", response)

	expected_response := "OK\n"
	if response == expected_response {
		fmt.Println("OK, server returned correct answer")
	} else {
		log.Fatalf("Not correct answer, expected %q, got %q", expected_response, response)
	}
}
