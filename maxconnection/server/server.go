package main

import (
	"os"
	"fmt"
	"net"
)

var usage = `Usage:
%v listenAddr`


func serve(conn net.Conn) {
	defer conn.Close()

}

func main() {
	arg := os.Args

	if len(arg) < 2 {
		fmt.Printf(usage, arg[0])
		return
	}

	listener, err := net.Listen("tcp", arg[1])
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go serve(conn)
	}
}
