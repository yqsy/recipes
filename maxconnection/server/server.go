package main

import (
	"os"
	"fmt"
	"net"
	"io"
	"strconv"
)

var usage = `Usage:
%v listenAddr limitSessionNum
`

type Context struct {
	conn net.Conn
}

type Global struct {
	acceptedNum int

	limitSessionNum int
}

func serve(ctx *Context) {
	defer ctx.conn.Close()
	io.Copy(ctx.conn, ctx.conn)
}

func main() {
	arg := os.Args

	if len(arg) < 3 {
		fmt.Printf(usage, arg[0])
		return
	}

	var err error
	gb := &Global{}
	gb.limitSessionNum, err = strconv.Atoi(arg[2])
	if err != nil {
		panic(err)
	}

	listener, err := net.Listen("tcp", arg[1])
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	for {
		ctx := &Context{}
		ctx.conn, err = listener.Accept()

		gb.acceptedNum ++

		if err != nil {
			fmt.Printf("current accepted sessionNum: %v\n", gb.acceptedNum)
			panic(err)
		}

		if gb.acceptedNum > gb.limitSessionNum {
			gb.acceptedNum --
			ctx.conn.Close()
			continue
		}

		go serve(ctx)
	}
}
