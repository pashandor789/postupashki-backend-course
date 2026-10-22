package main

import (
	"fmt"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		_, err = fmt.Fprint(conn, "OK\n")
		if err != nil {
			fmt.Println(err)
		}
		conn.Close()
	}

}
