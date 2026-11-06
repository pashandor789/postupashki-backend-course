package main

import (
	"net"
	"os"
)

func main() {
	l, err := net.Listen("tcp", ":8080")

	if err != nil {
		os.Exit(-1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		go func(conn net.Conn) {
			defer conn.Close()

			_, err = conn.Write([]byte("OK\n"))

			if err != nil {
				return
			}
		}(conn)
	}
}
