package main

import (
	"log"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	buf := make([]byte, 3)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}

	response := string(buf[:n])
	if response == "OK\n" {
		log.Println("Успех:", response)
	} else {
		log.Println("Неожиданный ответ:", response)
	}
}
