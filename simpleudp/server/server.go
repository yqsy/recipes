package main

import (
	"os"
	"fmt"
	"net"
)

var usage = `Usage:
%v listenAddr
`

func main() {

	arg := os.Args

	usage = fmt.Sprintf(usage, arg[0])

	if len(arg) < 2 {
		fmt.Printf(usage)
		return
	}

	serverAddr, err := net.ResolveUDPAddr("udp", arg[1])
	if err != nil {
		panic(err)
	}

	serverConn, err := net.ListenUDP("udp", serverAddr)

	if err != nil {
		panic(err)
	}

	buf := make([]byte, 2048)
	for {
		rn, remoteAddr, err := serverConn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}

		serverConn.WriteTo(buf[:rn], remoteAddr)
	}
}
