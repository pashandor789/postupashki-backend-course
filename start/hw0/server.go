package main 

import (
    "net"
    "log"
)

func main() {
    ln, err := net.Listen("tcp", ":8080")
    if err != nil {
	log.Fatal(err)
    }

    defer ln.Close()

    for {
	conn, err := ln.Accept()
	if err != nil {
	    log.Println(err)
	}

	msg := []byte("OK\n");
	_, err = conn.Write(msg)
	if err != nil {
	    log.Println(err)
	}

	conn.Close()
    }
}
