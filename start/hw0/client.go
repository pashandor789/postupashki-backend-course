package main

// Подключается к серверу на localhost:8080
// Читает ответ от сервера
// Проверяет, что получен ответ "OK\n"
// Корректно закрывает соединение
import (
	"bytes"
	"fmt"
	"log"
	"net"
)

func main() {
	sv, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	for {
		buf := make([]byte, 128)
		n, err := sv.Read(buf)
		if err != nil {
			fmt.Println("Ошибка!", err)
			continue
		}
		if bytes.Equal(buf[:n], []byte("OK\n")) {
			sv.Close()
			break
		}
	}
}
