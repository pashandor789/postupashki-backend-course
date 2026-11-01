package main

import (
	"fmt"
	"net"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	_, err := conn.Write([]byte("OK\n"))
	if err != nil {
		fmt.Println("Error writing to connection:", err)
		return
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	for {
		fmt.Println("Server is listening on port 8080...")
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			return
		}
		handleConnection(conn)
	}

}
