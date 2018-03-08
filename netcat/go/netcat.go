package main

import (
	"os"
	"fmt"
	"net"
	"io"
	"strconv"
)

func ifErrorExit(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func relay(conn net.Conn) {
	defer conn.Close()

	done := make(chan int, 2)

	active := 1
	passive := 2

	// [stdin] -> remote
	go func() {
		io.Copy(conn, os.Stdin)
		done <- active
	}()

	// stdout <- [remote]
	go func() {
		io.Copy(os.Stdout, conn)
		done <- passive
	}()

	first := <-done

	if first == active {
		conn.(*net.TCPConn).CloseWrite()
		<-done
	} else {
		// how to stop read from stdin?
		<-done
		conn.(*net.TCPConn).CloseWrite()
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
