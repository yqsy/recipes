package main

import (
	"os"
	"fmt"
	"net"
	"time"
	"encoding/binary"
)

var usage = `Usage:
%v listenAddr
`

type Context struct {
	conn net.Conn
}

func serve(ctx *Context) {
	defer ctx.conn.Close()
	now := time.Now()
	secs := now.Unix()

	err := binary.Write(ctx.conn, binary.BigEndian, secs)
	if err != nil {
		return
	}

	// safe close
	ctx.conn.(*net.TCPConn).CloseWrite()

	buf := make([]byte, 1024)
	for {
		rn, err := ctx.conn.Read(buf)
		_ = rn

		if err != nil {
			break
		}
	}
}

func main() {
	arg := os.Args

	usage = fmt.Sprintf(usage, arg[0])

	if len(arg) < 2 {
		fmt.Printf(usage)
		return
	}

	listener, err := net.Listen("tcp", arg[1])
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	for {
		ctx := &Context{}
		ctx.conn, err = listener.Accept()
		if err != nil {
			panic(err)
		}
		go serve(ctx)

	}
}
