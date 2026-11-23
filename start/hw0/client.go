package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

func ReadResponse(c net.Conn, limit int) error {
	r := io.LimitReader(c, int64(limit))
	b, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("Ошибка чтения: %v", err)
	}
	response := string(b)

	if response != "OK\n" {
		return fmt.Errorf("Вернулось значение %q, ожидалось %q\n", response, "OK\n")
	}
	fmt.Print(response)
	return err
}

func RunClient(host string) error {
	c, err := net.Dial("tcp", host)

	if err != nil {
		return fmt.Errorf("Ошибка соединения: %v", err)
	}

	defer c.Close()

	err = c.SetReadDeadline(time.Now().Add(time.Second * 10))

	if err != nil {
		return fmt.Errorf("Ошибка чтения: %v", err)
	}

	err = ReadResponse(c, 64)

	return err
}

func main() {

	err := RunClient("localhost:8080")

	if err != nil {
		log.Fatalf("Выполнение закончилось с ошибкой %v", err)
	}

}
