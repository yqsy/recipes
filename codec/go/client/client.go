package main

import (
	"net"
	"os"
	"fmt"
	"log"
)

var usage = `Usage:
%v connectAddr
`

type Context struct {
	conn net.Conn
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

	ctx := &Context{}

	var err error
	ctx.conn, err = net.Dial("tcp", arg[1])
	if err != nil {
		panic(err)
	}

	serve(ctx)
}
