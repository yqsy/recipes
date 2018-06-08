package main

import (
	"os"
	"fmt"
	"net"
	"io"
)

var (
	HalfCloseEnable = false
)

func relay(conn net.Conn) {
	defer conn.Close()

	done := make(chan int, 2)

	const active = 1
	const passive = 2

	// [stdin] -> remote
	go func(conn net.Conn, done chan int) {
		io.Copy(conn, os.Stdin)
		done <- active
	}(conn, done)

	// stdout <- [remote]
	go func(conn net.Conn, done chan int) {
		io.Copy(os.Stdout, conn)
		done <- passive
	}(conn, done)

	first := <-done

	if first == active {
		conn.(*net.TCPConn).CloseWrite()
		<-done
	} else {
		if HalfCloseEnable {
			<-done
			conn.(*net.TCPConn).CloseWrite()
		}
	}
}

func main() {

	arg := os.Args

	if len(arg) < 3 {
		fmt.Printf("Usage:\n  %v -l port\n  %v host port\n", arg[0], arg[0])
		return
	}

	if len(arg) > 3 && arg[3] == "-h" {
		HalfCloseEnable = true
	}

	if arg[1] == "-l" {
		//server
		addr := ":" + arg[2]
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			panic(err)
		}
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		listener.Close()
		relay(conn)
	} else {
		//client
		addr := arg[1] + ":" + arg[2]
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			panic(err)
		}
		relay(conn)
	}
}
