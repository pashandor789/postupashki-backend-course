package main

import (
	"log"
	"net"
)

func main() {
	l, err := net.Listen("tcp", "localhost:8080")

	if err != nil {
		log.Fatalf("Ошибка листенера: %v\n", err)
	}

	defer l.Close()

	for {
		conn, err := l.Accept()

		if err != nil {
			log.Fatalf("Ошибка подключения клиента: %v\n", err)
		}

		go func(c net.Conn) {
			defer c.Close()
			_, err := c.Write([]byte("OK\n"))

			if err != nil {
				log.Printf("Ошибка отправки ответа %v\n", err)
			}
		}(conn)
	}
}
