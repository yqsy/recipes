package main

import (
	"net"
	"fmt"
	"os"
	"bufio"
)

func dispatch(localConn net.Conn) {
	defer localConn.Close()

	bufReader := bufio.NewReader(localConn)
	firstByte, err := bufReader.Peek(1)
	if err != nil || len(firstByte) != 1 {
		return
	}

	if firstByte[0] == 0x04 {
		socksHandle4(localConn)
	} else if firstByte[0] == 0x05 {
		socksHandle5(localConn)
	}
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

		go dispatch(localConn)
	}
}
