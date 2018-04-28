package main

import (
	"os"
	"fmt"
	"net"
	"strconv"
)

var usage = `Usage:
%v listenAddr bufSize
`

type Context struct {
	bufSize int
	buf     []byte
	conn    net.Conn
}

func serve(ctx *Context) {
	defer ctx.conn.Close()

	for {
		rn, err := ctx.conn.Read(ctx.buf)
		if err != nil {
			return
		}

		wn, err := ctx.conn.Write(ctx.buf[:rn])
		if err != nil {
			return
		}

		_ = wn
	}
}

func main() {
	arg := os.Args

	usage = fmt.Sprintf(usage, arg[0])

	if len(arg) < 3 {
		fmt.Printf(usage)
		return
	}

	bufSize, err := strconv.Atoi(arg[2])
	if err != nil {
		fmt.Printf(usage)
		return
	}

	listener, err := net.Listen("tcp", arg[1])
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	for {
		ctx := &Context{bufSize: bufSize, buf: make([]byte, bufSize)}
		ctx.conn, err = listener.Accept()
		if err != nil {
			panic(err)
		}
		go serve(ctx)
	}
}
