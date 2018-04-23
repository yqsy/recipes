package main

import (
	"os"
	"fmt"
	"net"
	"log"
	"time"
	"github.com/yqsy/recipes/multiplexer/common"
)

var usage = `Usage:
%v listenip:listenport`

func serverChannelConnect(context *common.Context) {
	common.ServerChannelPassive(context, context.DmuxConnectAddr)
	log.Printf("read EOF from channel , reaccept channel")
}

func serverChannelBind(context *common.Context) {
	common.ServerChannelActive(context)
	log.Printf("read EOF from channel , reaccept channel")
}

func serverLocalListener(context *common.Context) {
	for {
		conn, err := context.DmuxLolcalListener.Accept()

		// TODO: 因为channel的关闭会导致listener的关闭,所以我暂时没做描述符满的操作(怎么区别两种关闭?)
		if err != nil {
			break
		}

		session := common.NewSession(conn)
		go common.ServeSessionActive(context, session)
	}

	log.Printf("listener close")
}

func doConnectWay(channelConn net.Conn, cmd common.ChannelBody) {
	context := common.NewContext(common.Connect, channelConn)
	var err error
	context.DmuxConnectAddr, err = cmd.GetConnectAddr()
	if err != nil {
		log.Printf("CONNECT addr error: %v", err)
		connectAckPack := common.NewConnectAckPack(false).Serialize()
		channelConn.Write(connectAckPack)
	} else {
		connectAckPack := common.NewConnectAckPack(true).Serialize()
		channelConn.Write(connectAckPack)
		log.Printf("CONNECT ok %v", channelConn.RemoteAddr())
		serverChannelConnect(context)
	}
}

func doBindWay(channelConn net.Conn, cmd common.ChannelBody) {
	context := common.NewContext(common.Connect, channelConn)
	var err error
	bindAddr, _ := cmd.GetBindAddr()

	context.DmuxLolcalListener, err = net.Listen("tcp", bindAddr)
	if err != nil {
		log.Printf("BIND addr error: %v", err)
		bindAckPack := common.NewBindAckPack(false).Serialize()
		channelConn.Write(bindAckPack)
	} else {
		bindAckPack := common.NewBindAckPack(true).Serialize()
		channelConn.Write(bindAckPack)
		log.Printf("BIND ok %v", channelConn.RemoteAddr())

		go serverLocalListener(context)

		serverChannelBind(context)

		// 在这里关闭,保证重启channel时能listen成功
		context.DmuxLolcalListener.Close()
	}
}

func main() {
	arg := os.Args
	usage = fmt.Sprintf(usage, arg[0])

	if len(arg) < 2 {
		fmt.Println(usage)
		return
	}

	listener, err := net.Listen("tcp", arg[1])

	if err != nil {
		panic(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept error %v", err)
			time.Sleep(time.Second * 3)
			continue
		}

		go func(channelConn net.Conn) {
			defer channelConn.Close()

			cmd, err := common.ReadChannelCmd(channelConn)

			if err != nil {
				log.Printf("recv err cmd len: %v", len(cmd))
				return
			}

			if cmd.IsConnect() {
				doConnectWay(channelConn, cmd)
			} else if cmd.IsBind() {
				doBindWay(channelConn, cmd)
			} else {
				log.Printf("un sopported cmd len: %v", len(cmd))
			}
		}(conn)
	}
}
