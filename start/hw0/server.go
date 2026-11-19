package main

import (
	"fmt"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Port are opening:", listener.Addr().String())

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			return
		}

		fmt.Println("Client connected:", conn.RemoteAddr().String())

		data, err := conn.Write([]byte("OK\n"))
		conn.Close()
		if err != nil {
			fmt.Println("Error writing to connection:", err)
			continue
		}
		fmt.Println("Data written:", data)

	}
}
