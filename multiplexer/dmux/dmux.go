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

func serverChannelConnect(ctx *common.Context) {
	common.ServerChannelPassive(ctx, ctx.DmuxConnectAddr)
	log.Printf("read EOF from channel , reaccept channel\n")
}

func serverChannelBind(ctx *common.Context) {
	common.ServerChannelActive(ctx)
	log.Printf("read EOF from channel , reaccept channel\n")
}

func serverLocalListener(ctx *common.Context) {
	for {
		conn, err := ctx.DmuxLolcalListener.Accept()

		// TODO: 因为channel的关闭会导致listener的关闭,所以我暂时没做描述符满的操作(怎么区别两种关闭?)
		if err != nil {
			break
		}

		session := common.NewSession(conn)
		go common.ServeSessionActive(ctx, session)
	}

	log.Printf("listener close\n")
}

func doConnectWay(channelConn net.Conn, cmd common.ChannelBody) {
	ctx := common.NewContext(common.Connect, channelConn)
	var err error
	ctx.DmuxConnectAddr, err = cmd.GetConnectAddr()
	if err != nil {
		log.Printf("CONNECT addr error: %v\n", err)
		connectAckPack := common.NewConnectAckPack(false).Serialize()
		channelConn.Write(connectAckPack)
	} else {
		connectAckPack := common.NewConnectAckPack(true).Serialize()
		channelConn.Write(connectAckPack)
		log.Printf("CONNECT ok %v\n", channelConn.RemoteAddr())
		serverChannelConnect(ctx)
	}
}

func doBindWay(channelConn net.Conn, cmd common.ChannelBody) {
	ctx := common.NewContext(common.Connect, channelConn)
	var err error
	bindAddr, _ := cmd.GetBindAddr()

	ctx.DmuxLolcalListener, err = net.Listen("tcp", bindAddr)
	if err != nil {
		log.Printf("BIND addr error: %v\n", err)
		bindAckPack := common.NewBindAckPack(false).Serialize()
		channelConn.Write(bindAckPack)
	} else {
		bindAckPack := common.NewBindAckPack(true).Serialize()
		channelConn.Write(bindAckPack)
		log.Printf("BIND ok %v\n", channelConn.RemoteAddr())

		go serverLocalListener(ctx)

		serverChannelBind(ctx)

		// 在这里关闭,保证重启channel时能listen成功
		ctx.DmuxLolcalListener.Close()
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
			log.Printf("accept error %v\n", err)
			time.Sleep(time.Second * 3)
			continue
		}

		go func(channelConn net.Conn) {
			defer channelConn.Close()

			cmd, err := common.ReadChannelCmd(channelConn)

			if err != nil {
				log.Printf("recv err cmd len: %v\n", len(cmd))
				return
			}

			if cmd.IsConnect() {
				doConnectWay(channelConn, cmd)
			} else if cmd.IsBind() {
				doBindWay(channelConn, cmd)
			} else {
				log.Printf("un sopported cmd len: %v\n", len(cmd))
			}
		}(conn)
	}
}
