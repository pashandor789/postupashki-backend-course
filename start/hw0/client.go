package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

// should be in confing -> later rework
const (
	server = "localhost:8080"
)

func main() {
	conn, err := net.Dial("tcp", server)
	if err != nil {
		log.Fatalf("Faliled to connect to the server %s : %v", server, err)
		return
	}
	fmt.Println("Successful connect to the server %s : %v", server, err)
	defer conn.Close()

	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to get answer from the server: %v", err)
	}

	fmt.Printf("Got answer from server: %q\n", response)

	expected_response := "OK\n"
	if response == expected_response {
		fmt.Println("OK, server returned correct answer")
	} else {
		log.Fatalf("Not correct answer, expected %q, got %q", expected_response, response)
	}
}
