package main

import (
	"net"
	"os"
	"fmt"
	"encoding/binary"
	"time"
)

var usage = `Usage:
%v connectAddr
`

type Context struct {
	conn net.Conn
}

func serve(ctx *Context) {
	defer ctx.conn.Close()

	var secs int64
	err := binary.Read(ctx.conn, binary.BigEndian, &secs)
	if err != nil {
		panic(err)
	}

	tm := time.Unix(secs, 0)
	fmt.Println(tm)
}

func main() {
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
