package main

import (
	"log"
	"net"
)

const (
	server_addr = "localhost:8080"
)

func connection_handler(conn net.Conn) {
	defer conn.Close()

	_, err := conn.Write([]byte("OK\n"))
	if err != nil {
		log.Println("Failed to send data to client:", err)
		return
	}
	log.Println("Succesfully sent data to client")
	return
}

func main() {
	listener, err := net.Listen("tcp", server_addr)
	if err != nil {
		log.Fatalf("Can't start server: %v", err)
	}
	defer listener.Close()
	log.Printf("Server is set up and listening port: %s", server_addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Can't get connection:", err)
			continue
		}
		log.Println("New connection:", conn.RemoteAddr())

		go connection_handler(conn)
	}
}
