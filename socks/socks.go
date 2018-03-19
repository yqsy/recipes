package main

import (
	"net"
	"fmt"
	"os"
)

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

		go socksHandle5(localConn)
	}
}
