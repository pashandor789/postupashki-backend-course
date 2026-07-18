package main

import (
	"log"
	"net"
)

const (
	address = "localhost:8080"
	network = "tcp"
	ok      = "OK\n"
)

func main() {
	listener, err := net.Listen(network, address)
	if err != nil {
		log.Printf("Ошибка в подключении: %v\n", err)
		return
	}

	defer listener.Close()

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Printf("Ошибка соединения с клиентом: %v", err)
		}

		_, err = connection.Write([]byte(ok))
		if err != nil {
			log.Printf("Ошибка записи: %v", err)
			connection.Close()
			break
		}

		// bufferRead := make([]byte, 256)

		// n, err := connection.Read(bufferRead)
		// if err != nil {
		// 	log.Printf("Ошибка чтения: %v\n", err)
		// 	return
		// }

		// if string(bufferRead[0:n]) != ok {
		// 	fmt.Println("Ошибка: получен неизвестный ответ: ", string(bufferRead[:n]))
		// }
		// fmt.Println(string(bufferRead[:n]))

		// _, err = connection.Write([]byte(ok))
		// if err != nil {
		// 	panic(err)
		// }
	}
}
