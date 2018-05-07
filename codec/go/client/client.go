package main

import (
	"net"
	"os"
	"fmt"
	"log"
	"github.com/yqsy/recipes/codec/go/proto"
	"github.com/yqsy/recipes/codec/go/common"
	"bufio"
)

var usage = `Usage:
%v connectAddr
`

type Context struct {
	conn net.Conn
}

func serve(ctx *Context) {
	defer ctx.conn.Close()

	query := &codec.Query{}
	query.Question = "hot are you?"

	err := common.WriteAMessage(ctx.conn, query)
	if err != nil {
		panic("write packet error")
	}
	bufReader := bufio.NewReader(ctx.conn)

	message, err := common.ReadAMessage(bufReader)

	if err != nil {
		panic(err)
	}

	log.Printf("%v \n", message.(*codec.Answer).Answer)
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
