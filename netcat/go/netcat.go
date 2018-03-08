package main

import (
	"os"
	"fmt"
	"net"
	"strconv"
	"io"
)

func ifErrorExit(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func relay(conn net.Conn) {
	defer conn.Close()

	stinDone := make(chan bool)
	remoteDone := make(chan bool)

	// [stdin] -> remote
	go func() {
		io.Copy(conn, os.Stdin)
		stinDone <- true
	}()

	// stdout <- [remote]
	go func() {
		io.Copy(os.Stdout, conn)
		remoteDone <- true
	}()

	for i := 0; i < 2; i++ {
		select {
		case <-stinDone:
			conn.(*net.TCPConn).CloseWrite()
			fmt.Println("[stdin] -> remote shutdown")
		case <-remoteDone:
			fmt.Println("stdout <- [remote] shutdown")
		}
	}
}

func main() {

	arg := os.Args

	if len(arg) < 3 {
		fmt.Printf("Usage:\n  %v -l port\n  %v host port\n", arg[0], arg[0])
		return
	}

	port, err := strconv.Atoi(arg[2])
	ifErrorExit(err)

	if arg[1] == "-l" {
		addr := ":" + strconv.Itoa(port)
		listener, err := net.Listen("tcp", addr)
		ifErrorExit(err)
		conn, err := listener.Accept()
		ifErrorExit(err)
		listener.Close()
		relay(conn)
	} else {
		addr := arg[1] + ":" + strconv.Itoa(port)
		conn, err := net.Dial("tcp", addr)
		ifErrorExit(err)
		relay(conn)
	}
}
