package main

import (
	"os"
	"fmt"
	"net"
	"syscall"
	"encoding/binary"
	"log"
)

var usage = `Usage:
%v listenAddr`

const SO_ORIGINAL_DST = 80

func serve(conn net.Conn) {

	defer conn.Close()

	file, err := conn.(*net.TCPConn).File()
	if err != nil {
		return
	}

	fd := int(file.Fd())

	addr, err := syscall.GetsockoptIPv6Mreq(fd, syscall.IPPROTO_IP, SO_ORIGINAL_DST)

	ipv4 := net.IPv4(addr.Multiaddr[4],
		addr.Multiaddr[5],
		addr.Multiaddr[6],
		addr.Multiaddr[7]).String()

	port := binary.BigEndian.Uint16([]byte{addr.Multiaddr[2], addr.Multiaddr[3]})

	log.Printf("dst: %v:%v", ipv4, port)

}

type Context struct {
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
