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
%v -R localhost:5001::5000 dmux_ip:dmux_port

-L: local listen and connect to remote in channel
-R: remote listen and connect to local in channel`

func serverChannelConnect(context *common.Context) {
	common.ServerChannelActive(context)
	log.Printf("read EOF from channel , reconnect channel")
}

func serverChannelBind(context *common.Context) {
	common.ServerChannelPassive(context, context.MultiplexerConnectAddr)
	log.Printf("read EOF from channel , reconnect channel")
}

func serveLocalListener(context *common.Context) {
	for {
		conn, err := context.MultiplexerLocalListener.Accept()

		// TODO: 因为channel的关闭会导致listener的关闭,所以我暂时没做描述符满的操作(怎么区别两种关闭?)
		if err != nil {
			break
		}

		session := common.NewSession(conn)
		go common.ServeSessionActive(context, session)
	}

	log.Printf("listener close")
}

func doConnectWay(arg []string) {
	pair := strings.Split(arg[2], ":")
	localListenAddr := pair[0] + ":" + pair[1]
	remoteConnectAddr := pair[2] + ":" + pair[3]
	dmuxAddr := arg[3]

	for {
		channelConn, err := net.Dial("tcp", dmuxAddr)

		if err != nil {
			log.Printf("dial error %v", dmuxAddr)
			time.Sleep(time.Second * 3)
			continue
		}

		connectPack := common.NewConnectPack(remoteConnectAddr).Serialize()
		wn, err := channelConn.Write(connectPack)
		if err != nil || wn != len(connectPack) {
			log.Printf("CONNECT error")
			continue
		}

		cmd, err := common.ReadChannelCmd(channelConn)
		if err != nil || !cmd.IsConnectOK() {
			log.Printf("CONNECT error")
			continue
		}

		log.Printf("CONNECT ok %v", dmuxAddr)

		context := common.NewContext(common.Connect, channelConn)
		context.MultiplexerLocalListener, err = net.Listen("tcp", localListenAddr)
		if err != nil {
			panic(err)
		}

		go serveLocalListener(context)

		serverChannelConnect(context)

		// 在这里关闭,保证重启channel时能listen成功
		context.MultiplexerLocalListener.Close()
	}
}

func doBindWay(arg []string) {
	pair := strings.Split(arg[2], ":")
	localConnectAddr := pair[0] + ":" + pair[1]
	remoteListenAddr := pair[2] + ":" + pair[3]
	dmuxAddr := arg[3]

	for {
		channelConn, err := net.Dial("tcp", dmuxAddr)

		if err != nil {
			log.Printf("dial error %v", dmuxAddr)
			time.Sleep(time.Second * 3)
			continue
		}

		bindPack := common.NewBindPack(remoteListenAddr).Serialize()
		wn, err := channelConn.Write(bindPack)
		if err != nil || wn != len(bindPack) {
			log.Printf("BIND error")
			continue
		}

		cmd, err := common.ReadChannelCmd(channelConn)
		if err != nil || !cmd.IsBindOK() {
			log.Printf("BIND error")
			continue
		}

		log.Printf("BIND ok %v", dmuxAddr)

		context := common.NewContext(common.Bind, channelConn)
		context.MultiplexerConnectAddr = localConnectAddr

		serverChannelBind(context)
	}
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
	if len(arg2) != 4 || len(arg3) != 2 {
		fmt.Println(usage)
		return
	}

	if arg[1] == "-L" {
		doConnectWay(arg)
	} else if arg[1] == "-R" {
		doBindWay(arg)
	} else {
		fmt.Println(usage)
	}

}
