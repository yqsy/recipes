package main

import (
	"log"
	"fmt"
	"os"
	"crypto/tls"
	"net"
	"io"
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
		conn.(*tls.Conn).CloseWrite()
		<-done
	} else {
		//how to stop read from stdin?
		//<-done
		//conn.(*tls.Conn).CloseWrite()
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	arg := os.Args

	if len(arg) < 3 {
		fmt.Printf("Usage:\n  %v -l port\n  %v host port\n", arg[0], arg[0])
		return
	}

	if arg[1] == "-l" {
		// server
		addr := ":" + arg[2]
		cer, err := tls.LoadX509KeyPair("fd.crt", "fd.key")
		if err != nil {
			panic(err)
		}
		config := &tls.Config{Certificates: []tls.Certificate{cer}}
		listener, err := tls.Listen("tcp", addr, config)
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
		// client
		addr := arg[1] + ":" + arg[2]
		config := &tls.Config{InsecureSkipVerify: true}
		conn, err := tls.Dial("tcp", addr, config)
		if err != nil {
			panic(err)
		}
		relay(conn)
	}
}
