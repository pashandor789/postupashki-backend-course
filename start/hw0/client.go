package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	// Подключаемся к серверу на localhost:8080
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка подключения к серверу:", err)
		os.Exit(1)
	}
	// Закрываем соединение в конце работы приложения
	defer conn.Close()

	// Читаем ответ от сервера до символа переноса строки
	message, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка чтения ответа:", err)
		os.Exit(1)
	}

	// Проверяем, что получен ответ "OK\n"
	if message == "OK\n" {
		fmt.Print("Ответ корректен: OK\n")
		os.Exit(0)
	} else {
		fmt.Fprintf(os.Stderr, "Некорректный ответ сервера: %q\n", message)
		os.Exit(1)
	}
}
