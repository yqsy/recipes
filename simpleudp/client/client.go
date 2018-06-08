package main

import (
	"os"
	"fmt"
	"net"
)

var usage = `Usage:
%v connectAddr
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

	localAddr, err := net.ResolveUDPAddr("udp", ":")
	if err != nil {
		panic(err)
	}

	conn, err := net.DialUDP("udp", localAddr, serverAddr)

	if err != nil {
		panic(err)
	}

	_, err = conn.Write([]byte("hello world"))
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 2048)
	rn, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%v\n", string(buf[:rn]))
}
