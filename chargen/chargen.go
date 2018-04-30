package main

import (
	"os"
	"fmt"
	"net"
)

var usage = `Usage:
%v listenAddr
`

type Context struct {
	conn net.Conn
}

type Chargen struct {
	line string

	currentIdx int
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func NewChargen() *Chargen {
	chargen := &Chargen{}

	for i := 33; i < 127; i++ {
		chargen.line = chargen.line + string(i)
	}

	return chargen
}

func (cg *Chargen) GetNext72Str() string {
	var str72 string

	end := min(cg.currentIdx+72, 94)

	str72 += cg.line[cg.currentIdx:end]

	if len(str72) < 72 {
		str72 += cg.line[0 : 72-len(str72)]
	}

	str72 += "\n"

	cg.currentIdx += 1
	if cg.currentIdx >= 94 {
		cg.currentIdx = 0
	}

	return str72
}

type Global struct {
	// chargen 发送的缓冲区,全局统一
	sendBuf []byte
}

func NewGlobal() *Global {

	chargen := NewChargen()

	gb := &Global{}

	var strall string
	for i := 0; i < 10; i++ {
		str72 := chargen.GetNext72Str()
		strall += str72
	}
	gb.sendBuf = []byte(strall)
	return gb
}

func serve(ctx *Context, gb *Global) {
	defer ctx.conn.Close()

	// discard all read
	go func(ctx *Context) {
		buf := make([]byte, 16*1024)
		for {
			rn, err := ctx.conn.Read(buf)
			_ = rn
			if err != nil {
				break
			}
		}
	}(ctx)

	for {
		wn, err := ctx.conn.Write(gb.sendBuf)
		if err != nil {
			break
		}
		_ = wn
	}
}

func main() {
	arg := os.Args

	usage = fmt.Sprintf(usage, arg[0])

	if len(arg) < 2 {
		fmt.Printf(usage)
		return
	}

	gb := NewGlobal()

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
		go serve(ctx, gb)

	}
}
