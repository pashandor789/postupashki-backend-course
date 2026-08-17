package main

import (
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Ошибка подключения к серверу:", err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)

	n, err := conn.Read(buffer)

	if err != nil {
		fmt.Println("Ошибка при чтении ответа от сервера:", err)
		return
	}

	response := string(buffer[:n])
	if response == "OK\n" {
		fmt.Println("Соединение с сервером успешно установлено")
	} else {
		fmt.Println("Неверный ответ от сервера:", response)
	}
}
