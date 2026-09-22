package main

import (
	"fmt"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")

	if err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
		return
	}

	defer listener.Close()

	fmt.Println("Сервер запущен на порту 8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Ошибка при подключении клиента:", err)
			continue
		}

		conn.Write([]byte("OK\n"))
		conn.Close()
	}
}
