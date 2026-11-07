package main

import (
	"fmt"
	"net"
	"os"
)

func run() error {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		_, err = fmt.Fprint(conn, "OK\n")
		if err != nil {
			fmt.Println(err)
		}
		conn.Close()
	}
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
