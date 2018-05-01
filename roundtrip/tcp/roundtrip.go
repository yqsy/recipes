package main

import (
	"os"
	"fmt"
	"net"
	"time"
	"log"
	"github.com/yqsy/recipes/roundtrip/common"
	"encoding/binary"
)

var usage = `Usage:
%v -s listenAddr
%v connectAddr
`

func RunServer(conn net.Conn) {
	defer conn.Close()

	for {
		msg := &common.Message{}
		err := binary.Read(conn, binary.BigEndian, msg)
		if err != nil {
			log.Println(err)
			break
		}

		msg.T2 = time.Now().UnixNano() / 1000

		err = binary.Write(conn, binary.BigEndian, msg)
		if err != nil {
			log.Println(err)
			break
		}
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	arg := os.Args

	usage = fmt.Sprintf(usage, arg[0], arg[0])

	if len(arg) < 2 {
		fmt.Printf(usage)
		return
	} else if len(arg) == 2 {
		// client

		var err error
		ctx := &common.ClientContext{Ticker: time.NewTicker(time.Second),
			StopTickerFlag: make(chan struct{}),
			StopAllFlag: make(chan struct{})}
		ctx.Conn, err = net.Dial("tcp", arg[1])
		if err != nil {
			panic(err)
		}

		common.RunClient(ctx)

	} else {
		// server
		if arg[1] != "-s" {
			fmt.Printf(usage)
			return
		}

		listener, err := net.Listen("tcp", arg[2])
		if err != nil {
			panic(err)
		}

		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		listener.Close()

		RunServer(conn)
	}
}
