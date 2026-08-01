package main

import (
	"fmt"
	"log"
	"net"
)

func handleConn(conn net.Conn) {
	defer conn.Close()

	if _, err := fmt.Fprint(conn, "OK\n"); err != nil {
		log.Printf("failed to send response: %v", err)
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleConn(conn)
	}
}
