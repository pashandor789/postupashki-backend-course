package main

import (
	"fmt"
	"net"
	"os"
)

func handleConnection(conn net.Conn) {
	// Обязательно закрываем соединение после отправки ответа
	defer conn.Close()
	
	// Отправляем строку "OK\n"
	_, err := conn.Write([]byte("OK\n"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка отправки данных:", err)
	}
}

func main() {
	// Слушаем подключения на порту 8080
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка запуска сервера:", err)
		os.Exit(1)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Ошибка подключения:", err)
			continue
		}
		// Запуск в горутине позволяет корректно обрабатывать множественные подключения
		go handleConnection(conn)
	}
}

