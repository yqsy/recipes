package main

import (
	"log"
	"os"
	"fmt"
	"net"
	"github.com/yqsy/recipes/roundtrip/common"
	"time"
	"unsafe"
	"encoding/binary"
	"bytes"
)

var usage = `Usage:
%v -s listenAddr
%v connectAddr
`

func RunServer(conn *net.UDPConn) {
	defer conn.Close()

	buf := make([]byte, unsafe.Sizeof(common.Message{}))

	for {

		rn, remoteAddr, err := conn.ReadFromUDP(buf)

		if err != nil {
			break
		}

		if rn != int(unsafe.Sizeof(common.Message{})) {
			continue
		}

		bufReader := bytes.NewBuffer(buf)
		msg := &common.Message{}
		err = binary.Read(bufReader, binary.BigEndian, msg)
		if err != nil {
			break
		}

		msg.T2 = time.Now().UnixNano() / 1000

		var bufSender bytes.Buffer
		binary.Write(&bufSender, binary.BigEndian, msg)
		_, err = conn.WriteToUDP(bufSender.Bytes(), remoteAddr)
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
		ctx := &common.ClientContext{Ticker: time.NewTicker(time.Second),
			StopTickerFlag: make(chan struct{}),
			StopAllFlag: make(chan struct{})}

		ctx.Conn, err = net.Dial("udp", arg[1])

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

		udpAddr, err := net.ResolveUDPAddr("udp", arg[2])
		if err != nil {
			panic(err)
		}

		listener, err := net.ListenUDP("udp", udpAddr)
		if err != nil {
			panic(err)
		}

		RunServer(listener)
	}
}
