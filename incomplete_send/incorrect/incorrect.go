package main

import (
	"net"
	"os"
	"fmt"
	"io"
)

func main() {
	arg := os.Args

	if len(arg) != 3 {
		fmt.Printf("usage: %v listenaddr file\n", arg[0])
		return
	}

	listener, err := net.Listen("tcp", arg[1])

	if err != nil {
		panic(err)
	}

	defer listener.Close()

	conn, err := listener.Accept()

	if err != nil {
		panic(err)
	}

	defer conn.Close()

	f, err := os.OpenFile(arg[2], os.O_RDONLY, 0666)

	if err != nil {
		panic(err)
	}

	io.Copy(conn, f)
}
