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

func serveSession(context *common.Context, session *common.Session) {
	defer session.Conn.Close()

	var err error
	session.Id, err = context.IdGen.GetFreeId()
	if err != nil {
		log.Printf("session id not enough")
		return
	}
	defer func() {
		context.IdGen.ReleaseId(session.Id)
		log.Printf("release id: %v", session.Id)
	}()

	context.ConnectSessionDict.Append(session)
	defer context.ConnectSessionDict.Del(session.Id)

	// SYN
	synPack := common.NewSynPack(session.Id)
	context.SendQueue.Put(synPack)
	recvPack := session.SendQueue.Take().(*common.ChannelPack)

	if !recvPack.Head.IsCmd() || !recvPack.Body.IsSynOK() {
		log.Printf("[%v]session SYN error", session.Id)
		return
	}

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
	defer context.ChannelConn.Close()

	// send channel (block queue , event loop)
	go common.ChannelSendEventLoop(context)

	// channel -> session
	for {
		channelPack, err := common.ReadChannelPack(context.ChannelConn)

		if err != nil {
			break
		}

		common.DispatchToSession(context, channelPack)
	}

	context.SendQueue.Put(nil)
	context.ConnectSessionDict.FinAll()
	context.CloseCond <- struct{}{}

	for i := 0; i < 2; i++ {
		<-context.CloseCond
	}
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
		go serveSession(context, session)
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
