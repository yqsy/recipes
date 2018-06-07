package main

import (
	"os"
	"fmt"
	"net"
	"time"
	"io"
	"log"
)


func serverConn(conn net.Conn) {
	defer conn.Close()

	conn.(*net.TCPConn).SetKeepAlive(true)
	conn.(*net.TCPConn).SetKeepAlivePeriod(3 * time.Second)

	io.Copy(conn, conn)

	log.Printf("%v -> %v stoped\n", conn.RemoteAddr(), conn.LocalAddr())
}

func main() {

	arg := os.Args
	if len(arg) < 2 {
		fmt.Printf("Usage:\n %v listenaddr\nExample:\n %v :10001\n", arg[0], arg[0])
		return
	}

	listener, err := net.Listen("tcp", arg[1])
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	for {
		localConn, err := listener.Accept()
		if err != nil {
			continue
		}

		go serverConn(localConn)
	}
}
