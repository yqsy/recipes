package main

import (
	"net"
	"os"
	"fmt"
	"io"
)

func main() {
	arg := os.Args

	if len(arg) != 3 {
		fmt.Printf("usage: %v listenaddr file", arg[0])
		return
	}

	listener, err := net.Listen("tcp", arg[1])

	if err != nil {
		panic(err)
	}

	defer listener.Close()

	conn, err := listener.Accept()

	if err != nil {
		panic(err)
	}

	defer conn.Close()

	f, err := os.OpenFile(arg[2], os.O_RDONLY, 0666)

	if err != nil {
		panic(err)
	}

	io.Copy(conn, f)
	// 发送完毕之后,发送fin
	conn.(*net.TCPConn).CloseWrite()

	// 接收对方的fin
	for {
		buf := make([]byte, 1024)
		rn, err := conn.Read(buf)
		_ = rn
		if err != nil {
			// EOF
			break
		}
	}

	// 关闭套接字
}
