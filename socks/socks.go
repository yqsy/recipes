package main

import (
	"net"
	"fmt"
	"os"
	"bufio"
	"github.com/yqsy/recipes/socks/socks4"
	"github.com/yqsy/recipes/socks/socks5"
	"log"
)

func dispatch(localConn net.Conn) {
	defer localConn.Close()

	bufReader := bufio.NewReader(localConn)
	firstByte, err := bufReader.Peek(1)
	if err != nil || len(firstByte) != 1 {
		return
	}

	if firstByte[0] == 0x04 {
		socks4.Socks4Handle(localConn, bufReader)
	} else if firstByte[0] == 0x05 {
		socks5.Socks5Handle(localConn, bufReader)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	arg := os.Args
	if len(arg) < 2 {
		fmt.Printf("Usage:\n %v listenaddr\nExample:\n %v :1080\n", arg[0], arg[0])
		return
	}

	listener, err := net.Listen("tcp", arg[1])
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	for {
		localConn, err := listener.Accept()
		if err != nil {
			continue
		}

		go dispatch(localConn)
	}
}
