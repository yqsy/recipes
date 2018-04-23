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
	go func(context *common.Context, session *common.Session) {
		for {
			val := session.AckQueue.Take()

			if val == nil {
				break
			}

			recvPack := val.(*common.ChannelPack)
			if recvPack.Head.IsCmd() && recvPack.Body.IsAck() {
				ackBytes, err := recvPack.Body.GetAckBytes()
				if err == nil {
					session.SendWaterMask.DropMask(ackBytes)
				}
			}
		}

		log.Printf("[%v]session ack done", session.Id)

	}(context, session)

	// session <-block queue channel
	go func(context *common.Context, session *common.Session) {
		for {
			recvPack := session.SendQueue.Take().(*common.ChannelPack)

			if recvPack.Head.IsCmd() && recvPack.Body.IsFin() {
				break
			}

			session.Conn.Write(recvPack.Body)
			session.RecvWaterMask += uint32(len(recvPack.Body))
			if session.RecvWaterMask > common.ResumeWaterMask {
				ackPack := common.NewAckPack(session.Id, session.RecvWaterMask)
				context.SendQueue.Put(ackPack)
				session.RecvWaterMask = 0
			}
		}

		// half close
		session.Conn.(*net.TCPConn).CloseWrite()
		session.CloseCond <- struct{}{}
		log.Printf("[%v]session <- channel done", session.Id)
	}(context, session)

	// session (阻塞read)-> channel
	for {
		session.SendWaterMask.WaitUntilCanBeWrite()

		if session.ChannelIsClose {
			break
		}

		buf := make([]byte, 16*1024)
		rn, err := session.Conn.Read(buf)

		if err != nil {
			break
		}

		payloadPack := common.NewPayloadPack(session.Id, buf[:rn])
		context.SendQueue.Put(payloadPack)
		session.SendWaterMask.RiseMask(uint32(rn))
	}

	// half close
	finPack := common.NewFinPack(session.Id)
	context.SendQueue.Put(finPack)
	session.CloseCond <- struct{}{}
	log.Printf("[%v]session -> channel done", session.Id)

	// stop ACK deal
	session.AckQueue.Put(nil)

	// 两边都半关闭完,释放连接
	for i := 0; i < 2; i++ {
		<-session.CloseCond
	}

	log.Printf("[%v]session <-> channel done", session.Id)
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

func serverChannelConnect(context *common.Context) {
	defer context.ChannelConn.Close()

	// send channel (block queue , event loop)
	go func(context *common.Context) {
		for {

			val := context.SendQueue.Take()
			// 退出的接口
			if val == nil {
				break
			}
			sendPack := val.(*common.ChannelPack)
			sendBytes := sendPack.Serialize()
			wn, err := context.ChannelConn.Write(sendBytes)

			if err != nil || wn != len(sendBytes) {
				break
			}
		}

		context.CloseCond <- struct{}{}
	}(context)

	// channel -> session
	for {
		channelPack, err := common.ReadChannelPack(context.ChannelConn)

		if err != nil {
			break
		}
		session := context.ConnectSessionDict.Find(channelPack.Head.Id)

		if session == nil {
			var cmd string
			if len(channelPack.Body) < 20 {
				cmd = string(channelPack.Body)
			}
			log.Printf("can't find session id:%v cmd:%v body:%v", channelPack.Head.Id, channelPack.Head.IsCmd(), cmd)
			continue
		}

		if channelPack.Head.IsCmd() && channelPack.Body.IsAck() {
			session.AckQueue.Put(channelPack)
		} else {
			session.SendQueue.Put(channelPack)
		}

	}

	context.SendQueue.Put(nil)
	context.ConnectSessionDict.FinAll()
	context.CloseCond <- struct{}{}

	for i := 0; i < 2; i++ {
		<-context.CloseCond
	}
	log.Printf("read EOF from channel , reconnect channel")
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
