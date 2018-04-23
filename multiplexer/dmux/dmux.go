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

func serverSession(context *common.Context, session *common.Session) {
	defer session.Conn.Close()

	defer context.ConnectSessionDict.Del(session.Id)

	log.Printf("[%v]session <-> channel relay", session.Id)

	// session <-block queue channel ack处理
	go common.SessionAckEventLoop(context, session)

	// session <-block queue channel
	go common.SessionSendEventLoop(context, session)

	// session (阻塞read)-> channel
	common.SessionReadEventLoop(context, session)

	// 两边都半关闭完,释放连接
	for i := 0; i < 2; i++ {
		<-session.CloseCond
	}

	log.Printf("[%v]session <-> channel done", session.Id)
}

func serverChannelConnect(context *common.Context) {

	// send channel (block queue , event loop)
	go common.ChannelSendEventLoop(context)

	// channel -> session
	for {
		channelPack, err := common.ReadChannelPack(context.ChannelConn)

		if err != nil {
			break
		}

		if channelPack.Head.IsCmd() && channelPack.Body.IsSyn() {
			conn, err := net.Dial("tcp", context.DmuxConnectAddr)
			if err != nil {
				log.Printf("dial error %v", err)
				synAckPack := common.NewSynAckPack(channelPack.Head.Id, false).Serialize()
				context.ChannelConn.Write(synAckPack)
			} else {
				synAckPack := common.NewSynAckPack(channelPack.Head.Id, true).Serialize()
				context.ChannelConn.Write(synAckPack)

				session := common.NewSession(conn)
				session.Id = channelPack.Head.Id
				context.ConnectSessionDict.Append(session)
				go serverSession(context, session)
			}
		} else {
			common.DispatchToSession(context, channelPack)
		}
	}

	context.SendQueue.Put(nil)
	context.ConnectSessionDict.FinAll()
	context.CloseCond <- struct{}{}

	for i := 0; i < 2; i++ {
		<-context.CloseCond
	}

	log.Printf("read EOF from channel , reaccept channel")
}

func serverChannelBind(context *common.Context) {

}

func serverChannel(channelConn net.Conn) {
	defer channelConn.Close()

	cmd, err := common.ReadChannelCmd(channelConn)

	if err != nil {
		log.Printf("recv err cmd len: %v", len(cmd))
		return
	}

	if cmd.IsConnect() {
		// ssh -NL
		context := common.NewContext(common.Connect, channelConn)
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

	} else if cmd.IsBind() {
		// ssh -NR

	} else {
		log.Printf("un sopported cmd len: %v", len(cmd))
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

		go serverChannel(conn)
	}
}
