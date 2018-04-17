package main

import (
	"os"
	"fmt"
	"strings"
	"github.com/yqsy/recipes/multiplexer/common"
	"net"
	"log"
	"time"
)

var usage = `Usage:
%v -L :5000:localhost:5001 dmux_ip:dmux_port
%v -R :5000:localhost:5001 dmux_ip:dmux_port

-L: local listen and connect to remote in channel
-R: remote listen and connect to local in channel`

func serveLocal(context *common.Context) {

}

func serverChannel(context *common.Context) {

}

func doLocalWay(arg []string) {
	pair := strings.Split(arg[2], ":")
	localListenAddr := pair[0] + ":" + pair[1]
	remoteConnectAddr := pair[2] + ":" + pair[3]
	dmuxAddr := arg[3]

	localListener, err := net.Listen("tcp", localListenAddr)
	defer localListener.Close()

	if err != nil {
		panic(err)
	}

	for {
		conn, err := net.Dial("tcp", dmuxAddr)

		if err != nil {
			log.Printf("connect error %v", dmuxAddr)
			time.Sleep(time.Second * 3)
			continue
		}

		log.Printf("connect ok")

		connectSynPack := common.NewConnectSynPack()

		wn, err := conn.Write(connectSynPack)
		if err != nil || wn != len(connectSynPack) {
			log.Printf("send SYN error")
			continue
		}

		cmd, err := common.ReadCmdLine(conn)

		if err != nil {
			log.Printf("recv err: %v")
			continue
		}

		if !cmd.IsConnectOK() {
			log.Printf("connect error")
			continue
		}

		context := common.NewContext(common.Connect)
		context.Channel = conn

		go serveLocal(context)

		serverChannel(context)
	}
}

func doRemoteWay(arg []string) {

}

func main() {
	arg := os.Args
	usage = fmt.Sprintf(usage, arg[0], arg[0])

	if len(arg) < 4 {
		fmt.Println(usage)
		return
	}

	arg2 := strings.Split(arg[2], ":")
	arg3 := strings.Split(arg[3], ":")
	if len(arg2) != 3 || len(arg3) != 1 {
		fmt.Println(usage)
		return
	}

	if arg[1] == "-L" {
		doLocalWay(arg)
	} else if arg[1] == "-R" {
		doRemoteWay(arg)
	} else {
		fmt.Println(usage)
	}

}
