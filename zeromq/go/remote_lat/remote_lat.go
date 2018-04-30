package main

import (
	"net"
	"os"
	"fmt"
	"bufio"
	"strconv"
	"github.com/yqsy/recipes/zeromq/go/common"
	"encoding/binary"
	"time"
)

var usage = `Usage:
%v connectAddr msgSize roundTripCount
`

func serve(ctx *common.Context, gb *common.Global) {
	defer ctx.Conn.Close()

	start := time.Now()

	pack := &common.Pack{Len: int32(gb.MsgSize), Body: make([]byte, gb.MsgSize)}

	for i := 0; i < gb.RoundTripCount; i++ {
		err := binary.Write(ctx.Conn, binary.BigEndian, pack.Len)
		if err != nil {
			panic(err)
		}

		err = binary.Write(ctx.Conn, binary.BigEndian, pack.Body)
		if err != nil {
			panic(err)
		}

		pack, err := ctx.ReadPackage()

		if err != nil {
			panic(err)
		}

		if int(pack.Len) != gb.MsgSize {
			panic("len not equal")
		}
	}

	elapsed := time.Since(start)

	fmt.Printf("%v message bytes\n", gb.MsgSize)
	fmt.Printf("%v round-trips\n", gb.RoundTripCount)
	fmt.Printf("%v seconds\n", elapsed.Seconds())
	fmt.Printf("%.2f round-trips per second\n", float64(gb.RoundTripCount)/float64(elapsed.Seconds()))
	fmt.Printf("%.2f latency [us]\n", 1000000*float64(elapsed.Seconds())/float64(gb.RoundTripCount)/2)
	fmt.Printf("%.2f band width[MiB/s]\n", float64(gb.MsgSize*gb.RoundTripCount)/float64(elapsed.Seconds())/1024/1024)
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

	ctx := &common.Context{ReadBuf: make([]byte, gb.MsgSize)}

	ctx.Conn, err = net.Dial("tcp", arg[1])
	if err != nil {
		panic(err)
	}
	ctx.BufReader = bufio.NewReader(ctx.Conn)

	serve(ctx, gb)
}
