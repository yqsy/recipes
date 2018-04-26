package main

import (
	"os"
	"fmt"
	"net"
	"bufio"
	"github.com/yqsy/recipes/socks/socks4"
	"github.com/yqsy/recipes/socks/socks5"
	"github.com/yqsy/recipes/httpproxy/httpproxy"
)

const (
	MaxRead = 4096
)

func dispatch(localConn net.Conn) {
	defer localConn.Close()

	limitReader := httpproxy.NewLimitReader(localConn, MaxRead)

	bufReader := bufio.NewReader(limitReader)
	firstByte, err := bufReader.Peek(1)
	if err != nil || len(firstByte) != 1 {
		return
	}

	if firstByte[0] == 0x04 {
		limitReader.SetNoLimit()
		socks4.Socks4Handle(localConn, bufReader)
	} else if firstByte[0] == 0x05 {
		limitReader.SetNoLimit()
		socks5.Socks5Handle(localConn, bufReader)
	} else {
		bytes, err := bufReader.Peek(7)
		if err != nil {
			return
		}
		if string(bytes) == "CONNECT" {
			// https proxy
			httpproxy.HttpsProxyHandle(localConn, bufReader, limitReader)
		} else {
			// http proxy
			httpproxy.HttpProxyHandle(localConn, bufReader, limitReader)
		}
	}
}

func main() {
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
