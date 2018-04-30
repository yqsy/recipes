package main

import (
	"os"
	"fmt"
	"net"
	"time"
	"encoding/binary"
	"log"
)

var usage = `Usage:
%v -s listenAddr
%v connectAddr
`

type Message struct {
	T1 int64
	T2 int64
}

type ClientContext struct {
	ticker *time.Ticker

	conn net.Conn

	stopTickerFlag chan struct{}

	stopAllFlag chan struct{}
}

func runClient(ctx *ClientContext) {
	defer ctx.conn.Close()

	// send my time
	go func(ctx *ClientContext) {

		quit := false
		for {
			select {
			case <-ctx.ticker.C:
				msg := &Message{T1: time.Now().UnixNano() / 1000}
				err := binary.Write(ctx.conn, binary.BigEndian, msg)
				if err != nil {
					quit = true
					break
				}

			case <-ctx.stopTickerFlag:
				quit = true
				break
			}

			if quit {
				break
			}
		}

		log.Printf("send goroutine exit\n")
		ctx.stopAllFlag <- struct{}{}
	}(ctx)

	// recv and calc

	for {
		msg := &Message{}
		err := binary.Read(ctx.conn, binary.BigEndian, msg)
		if err != nil {
			break
		}

		T3 := time.Now().UnixNano() / 1000

		roundTripTime := T3 - msg.T1
		clockOffset := msg.T2 - (msg.T1+T3)/2

		log.Printf("round trip(us): %v clock error(us): %v", roundTripTime, clockOffset)
	}

	ctx.stopTickerFlag <- struct{}{}

	<-ctx.stopAllFlag
}

func runServer(conn net.Conn) {
	defer conn.Close()

	for {
		msg := &Message{}
		err := binary.Read(conn, binary.BigEndian, msg)
		if err != nil {
			break
		}

		msg.T2 = time.Now().UnixNano() / 1000

		err = binary.Write(conn, binary.BigEndian, msg)
		if err != nil {
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
		ctx := &ClientContext{ticker: time.NewTicker(time.Second),
			stopTickerFlag: make(chan struct{}),
			stopAllFlag: make(chan struct{})}
		ctx.conn, err = net.Dial("tcp", arg[1])
		if err != nil {
			panic(err)
		}

		runClient(ctx)

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

		runServer(conn)
	}
}
