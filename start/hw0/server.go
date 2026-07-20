package main

import (
	"fmt"
	"log"
	"net"
)

// Слушает подключения на порту 8080
// При подключении клиента отправляет строку "OK\n"
// Закрывает соединение после отправки ответа
// Должен корректно обрабатывать множественные подключения

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	for {
		l, err := ln.Accept()
		if err != nil {
			fmt.Println("Ошибка! ", err)
			continue
		}
		go func(conn net.Conn) {
			defer conn.Close()
			_, err := conn.Write([]byte("OK\n"))
			if err != nil {
				fmt.Println("Ошибка! ", err)
				return
			}
		}(l)
	}
}
