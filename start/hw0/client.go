package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", ":8080")

	if err != nil {
		fmt.Println(err)
	}

	defer conn.Close()

	reader := bufio.NewReader(conn)

	msg, err := reader.ReadString('\n')

	if msg == "OK\n" {
		fmt.Println("OK recieved")
	} else {
		fmt.Println(err)
	}

}
