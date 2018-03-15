package main

import (
	"net"
	"fmt"
	"encoding/binary"
	"bufio"
	"log"
	"io"
	"os"
)

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

type Socks4Req struct {
	Version     byte
	CommandCode byte
	Port        uint16
	Ipv4Addr    [4]byte
}

func (socks4Req *Socks4Req) getIP() string {
	return fmt.Sprintf("%v.%v.%v.%v",
		socks4Req.Ipv4Addr[0],
		socks4Req.Ipv4Addr[1],
		socks4Req.Ipv4Addr[2],
		socks4Req.Ipv4Addr[3])
}

func (socks4Req *Socks4Req) getPort() string {
	return fmt.Sprintf("%v", socks4Req.Port)
}

func (socks4Req *Socks4Req) isSocks4a() bool {
	return socks4Req.Ipv4Addr[0] == 0 &&
		socks4Req.Ipv4Addr[1] == 0 &&
		socks4Req.Ipv4Addr[2] == 0 &&
		socks4Req.Ipv4Addr[3] != 0
}

func (socks4Req *Socks4Req) checkLegal(remoteAddr net.Addr) bool {
	// must be 0x04 for this version
	if socks4Req.Version != 0x04 {
		log.Printf("illegal Version: %v from: %v\n", socks4Req.Version, remoteAddr)
		return false
	}

	// 0x01 = establish a TCP/IP stream connection
	if socks4Req.CommandCode != 1 {
		log.Printf("illegal CommandCode: %v: from: %v\n", socks4Req.CommandCode, remoteAddr)
		return false
	}
	return true
}

type Socks4Res struct {
	NullByte byte
	States   byte
	Port     uint16
	Ipv4Addr [4]byte
}

func relayTcpUntilDie(localConn net.Conn, remoteAddr string, remoteConn net.Conn) {
	log.Printf("relay: %v <-> %v\n", localConn.RemoteAddr(), remoteAddr)
	done := make(chan bool, 2)

	go func(remoteConn net.Conn, localConn net.Conn, remoteAddr string, done chan bool) {
		io.Copy(remoteConn, localConn)
		remoteConn.(*net.TCPConn).CloseWrite()
		log.Printf("done: %v -> %v\n", localConn.RemoteAddr(), remoteAddr)
		done <- true
	}(remoteConn, localConn, remoteAddr, done)

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

func socksHandle(localConn net.Conn) {
	defer localConn.Close()

	var socks4Req Socks4Req

	// read 8 bytes header
	err := binary.Read(localConn, binary.BigEndian, &socks4Req)
	if err != nil {
		return
	}

	if !socks4Req.checkLegal(localConn.RemoteAddr()) {
		return
	}

	bufReader := bufio.NewReader(localConn)
	userNameBytes, err := bufReader.ReadSlice(0)

	if err != nil {
		return
	}
	// do not use
	_ = userNameBytes

	var remoteAddr string

	if socks4Req.isSocks4a() {
		domainBytes, err := bufReader.ReadSlice(0)
		if err != nil {
			return
		}
		remoteAddr = string(domainBytes[:len(domainBytes)-1]) + ":" + socks4Req.getPort()
	} else {
		remoteAddr = socks4Req.getIP() + ":" + socks4Req.getPort()
	}

	remoteConn, err := net.Dial("tcp", remoteAddr)

	if err != nil {
		return
	}

	defer remoteConn.Close()

	// write ack to client
	socks4Res := Socks4Res{
		0x00,
		0x5A,
		0x00,
		[4]byte{0x00, 0x00, 0x00, 0x00}}

	err = binary.Write(localConn, binary.BigEndian, &socks4Res)
	if err != nil {
		return
	}

	relayTcpUntilDie(localConn, remoteAddr, remoteConn)
}

func main() {
	arg := os.Args
	if len(arg) < 2 {
		fmt.Printf("Usage:\n %v listenaddr\nExample:\n %v :1080\n", arg[0], arg[0])
		return
	}

	listener, err := net.Listen("tcp", arg[1])
	panicOnError(err)

	defer listener.Close()

	for {
		localConn, err := listener.Accept()
		if err != nil {
			continue
		}

		go socksHandle(localConn)
	}
}
