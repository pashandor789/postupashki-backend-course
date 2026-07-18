package main

import (
	"fmt"
	"log"
	"net"
)

const (
	address = "localhost:8080"
	network = "tcp"
	ok      = "OK\n"
)

func main() {
	connection, err := net.Dial(network, address)
	if err != nil {
		log.Printf("Ошибка соединения: %v", err)
		return
	}
	defer connection.Close()

	for {
		buf := make([]byte, 256)
		n, err := connection.Read(buf)
		if err != nil {
			log.Printf("Ошибка чтения: %v", err)
			return
		}

		if string(buf[0:n]) == ok {
			fmt.Println(string(buf[0:n]))
			break
		}
	}

}
