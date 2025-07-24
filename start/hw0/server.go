package main

import (
	"log"
	"net"
)

func handle_connection(conn net.Conn) {
	defer conn.Close()

	_, err := conn.Write([]byte("OK\n"))
	if err != nil {
		log.Println("Can't send data to client:", err)
		return
	}
	log.Println("OK was sent to client:", conn.RemoteAddr())
}

func main() {
	const addr = "localhost:8080"

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Can't start server: %v", err)
	}
	defer listener.Close()
	log.Printf("Server is set up and listening port: %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Can't get connection:", err)
			continue
		}
		log.Println("New connection:", conn.RemoteAddr())

		go handle_connection(conn)
	}
}
