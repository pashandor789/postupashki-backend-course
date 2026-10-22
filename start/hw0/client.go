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
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	answer, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		conn.Close()
		os.Exit(1)
	}
	if answer == "OK\n" {
		fmt.Printf("All good\n")
	} else {
		fmt.Printf("Incorrect answer\n")
		conn.Close()
		os.Exit(1)
	}

}
