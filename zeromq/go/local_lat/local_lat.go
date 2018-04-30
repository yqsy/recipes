package main

import (
	"os"
	"fmt"
	"net"
	"bufio"
	"strconv"
	"github.com/yqsy/recipes/zeromq/go/common"
	"encoding/binary"
)

var usage = `Usage:
%v listenAddr msgSize roundTripCount
`

func serve(ctx *common.Context, gb *common.Global) {
	defer ctx.Conn.Close()

	for i := 0; i < gb.RoundTripCount; i++ {
		pack, err := ctx.ReadPackage()

		if err != nil {
			panic(err)
		}

		if int(pack.Len) != gb.MsgSize {
			panic("len not equal")
		}

		err = binary.Write(ctx.Conn, binary.BigEndian, pack.Len)
		if err != nil {
			panic(err)
		}

		err = binary.Write(ctx.Conn, binary.BigEndian, pack.Body)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	arg := os.Args
	usage = fmt.Sprintf(usage, arg[0])

	if len(arg) < 4 {
		fmt.Printf(usage)
		return
	}

	var err error
	gb := &common.Global{}
	gb.MsgSize, err = strconv.Atoi(arg[2])
	if err != nil {
		panic(err)
	}

	gb.RoundTripCount, err = strconv.Atoi(arg[3])
	if err != nil {
		panic(err)
	}

	listener, err := net.Listen("tcp", arg[1])
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	for {
		ctx := &common.Context{ReadBuf: make([]byte, gb.MsgSize)}
		ctx.Conn, err = listener.Accept()
		ctx.BufReader = bufio.NewReader(ctx.Conn)

		if err != nil {
			panic(err)
		}
		go serve(ctx, gb)

	}
}
