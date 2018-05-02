package main

import (
	"net"
	"os"
	"fmt"
	"log"
	"github.com/yqsy/recipes/chat/common"
)

var usage = `Usage:
%v connectAddr
`

type Context struct {
	conn net.Conn
}

func serve(ctx *Context) {
	defer ctx.conn.Close()

	// stdin -> remote
	go func(ctx *Context) {
		buf := make([]byte, 16384)

		for {
			rn, err := os.Stdin.Read(buf)
			if err != nil {
				break
			}

			common.WriteToConnRaw(ctx.conn, buf[:rn])
		}
	}(ctx)

	// stdout <- remote
	for {
		msg, err := common.ReadFromConn(ctx.conn)

		if err != nil {
			break
		}

		fmt.Print(string(msg.Body))
	}

	log.Printf("client exit\n")
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
