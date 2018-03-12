package main

import (
	"os"
	"fmt"
	"net"
	"log"
	"io"
)

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func relay(localConn net.Conn, remoteAddr string) {
	defer localConn.Close()

	remoteConn, err := net.Dial("tcp", remoteAddr)

	if err != nil {
		log.Printf("connect err: %v -> %v\n", localConn.RemoteAddr(), remoteAddr)
		return
	}

	defer remoteConn.Close()
	log.Printf("relay: %v <-> %v\n", localConn.RemoteAddr(), remoteConn.RemoteAddr())

	done := make(chan bool, 2)

	go func(remoteConn net.Conn, localConn net.Conn, done chan bool) {
		io.Copy(remoteConn, localConn)
		remoteConn.(*net.TCPConn).CloseWrite()
		log.Printf("done: %v -> %v\n", localConn.RemoteAddr(), remoteConn.RemoteAddr())
		done <- true
	}(remoteConn, localConn, done)

	go func(localConn net.Conn, remoteConn net.Conn, done chan bool) {
		io.Copy(localConn, remoteConn)
		localConn.(*net.TCPConn).CloseWrite()
		log.Printf("done: %v <- %v\n", localConn.RemoteAddr(), remoteConn.RemoteAddr())
		done <- true
	}(localConn, remoteConn, done)

	for i := 0; i < 2; i++ {
		<-done
	}
}

func main() {
	arg := os.Args

	if len(arg) < 4 {
		fmt.Printf("Usage:\n %v listenaddr listenport transimitaddr transimitport\n"+
			"example:\n %v 0.0.0.0 10000 localhost 20000\n", arg[0], arg[0])
		return
	}

	listenAddr := arg[1] + ":" + arg[2]

	listener, err := net.Listen("tcp", listenAddr)
	panicOnError(err)

	remoteAddr := arg[3] + ":" + arg[4]

	for {
		localConn, err := listener.Accept()

		if err != nil {
			continue
		}

		go relay(localConn, remoteAddr)
	}
}
