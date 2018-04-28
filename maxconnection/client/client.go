package main

import (
	"net"
	"os"
	"fmt"
	"strconv"
	"sync/atomic"
	"io"
	"time"
)

var usage = `Usage:
%v connectAddr sessionNum 
`

type Context struct {
	conn net.Conn
}

type Global struct {
	maxSession int

	connectedNum int64

	allSession []*Context

	start time.Time

	closeSignal chan struct{}
}

func (gb *Global) addSession(ctx *Context) {
	gb.allSession = append(gb.allSession, ctx)
}

func serve(ctx *Context, gb *Global) {
	defer ctx.conn.Close()

	if atomic.AddInt64(&gb.connectedNum, 1) == int64(gb.maxSession) {
		fmt.Printf("%v session connected cost: %v s\n", gb.maxSession, time.Since(gb.start).Seconds())
	}

	io.Copy(ctx.conn, ctx.conn)

	curConnectedNum := atomic.AddInt64(&gb.connectedNum, -1)

	if curConnectedNum == 0 {
		fmt.Printf("all session closed\n")
		gb.closeSignal <- struct{}{}
	} else {
		fmt.Printf("%v closed, current num: %v\n", ctx.conn.LocalAddr(), curConnectedNum)
	}
}

func main() {
	arg := os.Args

	if len(arg) < 3 {
		fmt.Printf(usage, arg[0])
		return
	}

	var err error
	gb := &Global{closeSignal: make(chan struct{}), start: time.Now()}
	gb.maxSession, err = strconv.Atoi(arg[2])

	for i := 0; i < gb.maxSession; i++ {
		ctx := &Context{}
		ctx.conn, err = net.Dial("tcp", arg[1])
		if err != nil {
			fmt.Printf("current connected sessionNum: %v\n", atomic.LoadInt64(&gb.connectedNum))
			panic(err)
		}

		gb.addSession(ctx)
		go serve(ctx, gb)
	}

	<-gb.closeSignal
}
