package main

import (
	"net"
	"time"
	"sync/atomic"
	"fmt"
	"sync"
	"strconv"
	"os"
)

var usage = `Usage:
%v connectAddr bufSize sessionNum timeout`

type Context struct {
	bufSize   int
	buf       []byte
	timeOut   int
	conn      net.Conn
	readBytes int64
	readMsgs  int64
}

type Global struct {
	connectedNum int64
	sessionNum   int
	allCtx       []*Context

	// protect slice
	mtx sync.Mutex

	closeSignal chan struct{}
}

func (gb *Global) cumulateReadState() (bytes int64, msgs int64) {
	gb.mtx.Lock()
	defer gb.mtx.Unlock()

	var allReadBytes int64
	for _, ctx := range gb.allCtx {
		allReadBytes += ctx.readBytes
	}

	var allReadMsgs int64
	for _, ctx := range gb.allCtx {
		allReadMsgs += ctx.readMsgs
	}

	return allReadBytes, allReadMsgs
}

func (gb *Global) addCtx(ctx *Context) {
	gb.mtx.Lock()
	defer gb.mtx.Unlock()
	gb.allCtx = append(gb.allCtx, ctx)
}

func serve(ctx *Context, gb *Global) {
	// defer ctx.conn.Close()

	go func(ctx *Context, gb *Global) {
		time.Sleep(time.Duration(ctx.timeOut) * time.Second)

		ctx.conn.Close()

		if atomic.AddInt64(&gb.connectedNum, -1) == 0 {

			allReadBytes, allReadMsgs := gb.cumulateReadState()
			fmt.Printf("%v total bytes read\n", allReadBytes)
			fmt.Printf("%v total messages read\n", allReadMsgs)
			fmt.Printf("%.2f average message size\n", float64(allReadBytes)/float64(allReadMsgs))
			fmt.Printf("%.2f MiB/s throughput\n", float64(allReadBytes)/float64(ctx.timeOut)/1024/1024)

			gb.closeSignal <- struct{}{}
		}

	}(ctx, gb)

	if atomic.AddInt64(&gb.connectedNum, 1) == int64(gb.sessionNum) {
		fmt.Printf("all connected\n")
	}

	ctx.conn.Write(ctx.buf)

	for {
		rn, err := ctx.conn.Read(ctx.buf)
		if err != nil {
			return
		}

		wn, err := ctx.conn.Write(ctx.buf[:rn])
		if err != nil {
			return
		}

		ctx.readMsgs++
		ctx.readBytes += int64(rn)
		_ = wn
	}
}

func main() {
	arg := os.Args

	usage = fmt.Sprintf(usage, arg[0])

	if len(arg) < 5 {
		fmt.Printf(usage)
		return
	}

	var gb = &Global{closeSignal: make(chan struct{})}

	bufSize, err := strconv.Atoi(arg[2])
	if err != nil {
		fmt.Printf(usage)
		return
	}

	gb.sessionNum, err = strconv.Atoi(arg[3])
	if err != nil {
		fmt.Printf(usage)
		return
	}

	timeOut, err := strconv.Atoi(arg[4])
	if err != nil {
		fmt.Printf(usage)
		return
	}

	for i := 0; i < gb.sessionNum; i++ {
		ctx := &Context{bufSize: bufSize,
			buf: make([]byte, bufSize),
			timeOut: timeOut}
		ctx.conn, err = net.Dial("tcp", arg[1])

		if err != nil {
			panic(err)
		}

		gb.addCtx(ctx)
		go serve(ctx, gb)
	}

	<-gb.closeSignal
}
