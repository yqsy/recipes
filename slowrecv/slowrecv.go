package main

import (
	"os"
	"fmt"
	"net"
	"io"
)



func recv(conn net.Conn) {
	defer conn.Close()
	io.Copy(os.Stdout, conn)
}

func main() {

	arg := os.Args

	if len(arg) < 2 {
		fmt.Printf("Usage:\n  %v -l port\n ", arg[0])
		return
	}

	if arg[1] == "-l" {
		//server
		addr := ":" + arg[2]
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			panic(err)
		}
		defer listener.Close()

		for {
			conn, err := listener.Accept()

			if err != nil {
				continue
			}

			go recv(conn)
		}

	} else {
		fmt.Printf("Usage:\n  %v -l port\n ", arg[0])
	}
}
