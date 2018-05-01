package common

import (
	"time"
	"net"
	"encoding/binary"
	"log"
)

type Message struct {
	T1 int64
	T2 int64
}

type ClientContext struct {
	Ticker *time.Ticker

	Conn net.Conn

	StopTickerFlag chan struct{}

	StopAllFlag chan struct{}
}

func RunClient(ctx *ClientContext) {
	defer ctx.Conn.Close()

	// send my time
	go func(ctx *ClientContext) {

		quit := false
		for {
			select {
			case <-ctx.Ticker.C:
				msg := &Message{T1: time.Now().UnixNano() / 1000}
				err := binary.Write(ctx.Conn, binary.BigEndian, msg)
				if err != nil {
					quit = true
					break
				}

			case <-ctx.StopTickerFlag:
				quit = true
				break
			}

			if quit {
				break
			}
		}

		log.Printf("send goroutine exit\n")
		ctx.StopAllFlag <- struct{}{}
	}(ctx)

	// recv and calc

	for {
		msg := &Message{}
		err := binary.Read(ctx.Conn, binary.BigEndian, msg)
		if err != nil {
			break
		}

		T3 := time.Now().UnixNano() / 1000

		roundTripTime := T3 - msg.T1
		clockOffset := msg.T2 - (msg.T1+T3)/2

		log.Printf("round trip(us): %v clock error(us): %v", roundTripTime, clockOffset)
	}

	ctx.StopTickerFlag <- struct{}{}

	<-ctx.StopAllFlag
}

