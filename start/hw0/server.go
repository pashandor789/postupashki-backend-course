package main

import (
	"log"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		_, err = conn.Write([]byte("OK\n"))
		if err != nil {
			conn.Close()
			continue
		}

		conn.Close()

	}

}
