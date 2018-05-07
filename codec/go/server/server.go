package main

import (
	"os"
	"fmt"
	"net"
	"log"
	"github.com/golang/protobuf/proto"
	"github.com/yqsy/recipes/codec/go/proto"
	"bufio"
	"github.com/yqsy/recipes/codec/go/common"
	"reflect"
)

var usage = `Usage:
%v listenAddr
`

type Context struct {
	conn net.Conn
}

type Callback func(ctx *Context, message proto.Message)

type Global struct {
	callbacks map[interface{}]Callback
}

func fooQuery(ctx *Context, message proto.Message) {
	query := message.(*codec.Query)
	log.Printf("%v\n", query.Question)

	answer := &codec.Answer{}
	answer.Answer = "i m fine thank you, and you?"

	err := common.WriteAMessage(ctx.conn, answer)
	if err != nil {

		// error do nothing
		return
	}
}

func fooEmpty(ctx *Context, message proto.Message) {
	empty := message.(*codec.Empty)
	_ = empty
}

func serve(ctx *Context, gb *Global) {
	defer ctx.conn.Close()

	bufReader := bufio.NewReader(ctx.conn)

	for {
		message, err := common.ReadAMessage(bufReader)

		if err != nil {
			log.Printf("read message err: %v\n", err)
			break
		}

		if cb, ok := gb.callbacks[reflect.TypeOf(message)]; ok {
			cb(ctx, message)
		} else {
			log.Printf("callback didn't find\n")
			break
		}
	}

	log.Printf("remote: %v close\n", ctx.conn.RemoteAddr())
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

	gb := &Global{callbacks: make(map[interface{}]Callback)}
	gb.callbacks[reflect.TypeOf((*codec.Query)(nil))] = fooQuery
	gb.callbacks[reflect.TypeOf((*codec.Empty)(nil))] = fooEmpty

	for {
		ctx := &Context{}
		ctx.conn, err = listener.Accept()
		if err != nil {
			panic(err)
		}
		go serve(ctx, gb)
	}
}
