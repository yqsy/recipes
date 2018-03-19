package main

import (
	"net"
	"io"
	"log"
	"bufio"
)

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func relayTcpUntilDie(localConn net.Conn, remoteAddr string, remoteConn net.Conn, bufReader *bufio.Reader) {
	log.Printf("relay: %v <-> %v\n", localConn.RemoteAddr(), remoteAddr)
	done := make(chan bool, 2)

	go func(remoteConn net.Conn, localConn net.Conn, remoteAddr string, bufReader *bufio.Reader, done chan bool) {
		io.Copy(remoteConn, bufReader)
		remoteConn.(*net.TCPConn).CloseWrite()
		log.Printf("done: %v -> %v\n", localConn.RemoteAddr(), remoteAddr)
		done <- true
	}(remoteConn, localConn, remoteAddr, bufReader, done)

	go func(localConn net.Conn, remoteConn net.Conn, remoteAddr string, done chan bool) {
		io.Copy(localConn, remoteConn)
		localConn.(*net.TCPConn).CloseWrite()
		log.Printf("done: %v <- %v\n", localConn.RemoteAddr(), remoteAddr)
		done <- true
	}(localConn, remoteConn, remoteAddr, done)

	for i := 0; i < 2; i++ {
		<-done
	}
}
