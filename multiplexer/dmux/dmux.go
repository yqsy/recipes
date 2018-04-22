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

	// session <-block queue channel
	go func(context *common.Context, session *common.Session) {
		for {
			recvPack := session.SendQueue.Take().(*common.ChannelPack)

			if recvPack.Head.IsCmd() {
				if recvPack.Body.IsFin() {
					break
				} else if recvPack.Body.IsAck() {
					ackBytes, err := recvPack.Body.GetAckBytes()
					if err != nil {
						session.SendWaterMask.DropMask(ackBytes)
					}
				}
			}

			wn, err := session.Conn.Write(recvPack.Body)
			if err != nil || wn != len(recvPack.Body) {
				break
			}

			session.RecvWaterMask += uint32(wn)
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

		buf := make([]byte, 16*1024*1024)
		rn, err := session.Conn.Read(buf)

		if err != nil {
			break
		}

		payloadPack := common.NewPayloadPack(session.Id, buf[rn:])
		context.SendQueue.Put(payloadPack)
		session.SendWaterMask.RiseMask(uint32(rn))
	}

	// half close
	finPack := common.NewFinPack(session.Id)
	context.SendQueue.Put(finPack)
	session.CloseCond <- struct{}{}
	log.Printf("[%v]session -> channel done", session.Id)

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
				synAckPack := common.NewSynAckPack(false).Serialize()
				context.ChannelConn.Write(synAckPack)
			} else {
				synAckPack := common.NewSynAckPack(true).Serialize()
				context.ChannelConn.Write(synAckPack)

				session := common.NewSession(conn)
				session.Id = channelPack.Head.Id
				context.ConnectSessionDict.Append(session)
				go serverSession(context, session)
			}
		}

		session := context.ConnectSessionDict.Find(channelPack.Head.Id)
		session.SendQueue.Put(channelPack)
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
