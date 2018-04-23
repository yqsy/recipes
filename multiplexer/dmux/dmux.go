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

func serverChannelConnect(context *common.Context) {

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
