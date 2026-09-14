package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Printf("Error connecting to server %s\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	//read response
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("error reading response: %v\n", err)
		os.Exit(1)
	}

	if response == "OK\n" {
		fmt.Println("Recieved answer: OK")
	} else {
		fmt.Printf("recieved unexpected answer: %q (Expected OK) \n", response)
		os.Exit(1)
	}

}
