package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Printf("Server start error %v \n", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Server started at 8080")

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Error accepting connection", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	//send ok
	_, err := conn.Write([]byte("OK\n"))

	if err != nil {
		fmt.Printf("Error sending response to client %s: %v\n", conn.RemoteAddr(), err)
		return
	}

	fmt.Printf("Sent answer to client %s\n", conn.RemoteAddr())

}
