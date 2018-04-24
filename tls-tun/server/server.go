package main

import (
	"crypto/tls"
	"net"
	"bufio"
	"fmt"
	"os"
)

func handleConn(conn net.Conn) {
	defer conn.Close()

	bufReader := bufio.NewReader(conn)

	for {
		line, err := bufReader.ReadString('\n')
		if err != nil {
			break
		}

		fmt.Print(line)

		wn, err := conn.Write([]byte("world\n"))
		_ = wn
		if err != nil {
			break
		}
	}
}

func main() {
	arg := os.Args

	if len(arg) < 2 {
		fmt.Printf("Usage:\n%v listenAddr\n", arg[0])
		return
	}

	cer, err := tls.LoadX509KeyPair("fd.crt", "fd.key")
	if err != nil {
		panic(err)
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}

	listener, err := tls.Listen("tcp", arg[1], config)
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		go handleConn(conn)
	}
}
