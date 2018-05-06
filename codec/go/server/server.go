package main

import (
	"os"
	"fmt"
	"net"
	"log"
)

var usage = `Usage:
%v listenAddr
`

type Context struct {
	conn net.Conn
}

type Global struct {

}

func serve(ctx *Context) {
	defer ctx.conn.Close()

}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

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
