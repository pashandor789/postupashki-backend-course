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
		log.Printf("Ошибка в подключении: %v\n", err)
		return
	}

	defer connection.Close()

	for {
		bufferRead := make([]byte, 256)

		n, err := connection.Read(bufferRead)
		if err != nil {
			log.Printf("Ошибка чтения: %v\n", err)
			return
		}

		if string(bufferRead[:n]) != ok {
			fmt.Println("Ошибка: получен неизвестный ответ: ", string(bufferRead[:n]))
		}
		fmt.Println(string(bufferRead[:n]))

		// _, err = connection.Write([]byte(ok))
		// if err != nil {
		// 	panic(err)
		// }
	}
}
