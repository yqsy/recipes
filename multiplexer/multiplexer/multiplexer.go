package main

import (
	"os"
	"fmt"
	"strings"
	"github.com/yqsy/recipes/multiplexer/common"
	"net"
	"log"
	"time"
	"encoding/binary"
	"io"
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

	context.ConnectSessionDict.Append(session.Id, session)
	defer context.ConnectSessionDict.Del(session.Id)

	// SYN
	synMsg := common.NewMsg(session.Id, []byte("SYN"))
	context.SendQueue.Put(synMsg)
	val := session.SendQueue.Take()
	if val == nil || !val.(*common.ChannelBody).IsSynOK() {
		log.Printf("[%v]%v <-> dmux [dmux] SYN ERROR", session.Id, session.Conn.RemoteAddr())
		return
	}

	log.Printf("[%v]%v <-> dmux relay", session.Id, session.Conn.RemoteAddr())

	// 维护session <-(阻塞blockqueue) multiplexer
	go func(context *common.Context, session *common.Session) {
		for {
			val := session.SendQueue.Take()

			// FIN
			if val == nil {
				break
			}

			msg := val.(*common.Msg)

			wn, err := session.Conn.Write(msg.Data)
			if err != nil || wn != len(msg.Data) {
				break
			}

			session.RecvWaterMask += uint32(wn)
			if session.RecvWaterMask > common.ResumeWaterMask {
				ackPack := []byte(fmt.Sprintf("ACK %v", session.RecvWaterMask))
				ackMsg := common.NewMsg(session.Id, ackPack)
				context.SendQueue.Put(ackMsg)
				session.RecvWaterMask = 0
			}
		}

		// half close
		session.Conn.(*net.TCPConn).CloseWrite()
		session.CloseCond <- struct{}{}
		log.Printf("[%v]%v <- dmux done", session.Id, session.Conn.RemoteAddr())
	}(context, session)

	// session (阻塞read)-> multiplexer 投递到blockqueue中
	for {
		session.SendWaterMask.WaitUntilCanBeWrite()

		buf := make([]byte, 16*1024*1024)
		rn, err := session.Conn.Read(buf)

		if err != nil {
			break
		}

		payload := common.NewMsg(session.Id, buf[rn:])
		context.SendQueue.Put(payload)
		session.SendWaterMask.RiseMask(uint32(rn))
	}

	// half close
	finMsg := common.NewMsg(session.Id, nil)
	context.SendQueue.Put(finMsg)
	session.CloseCond <- struct{}{}
	log.Printf("[%v]%v -> dmux done", session.Id, session.Conn.RemoteAddr())

	// 两边都半关闭完,释放连接
	for i := 0; i < 2; i++ {
		<-session.CloseCond
	}
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
}

func serveChannel(context *common.Context) {
	defer context.Channel.Close()

	// 维护multiplexer -> channel (阻塞blockqueue)
	go func(context *common.Context) {

	}(context)

	// multiplexer <-(阻塞read) channel
	for {
		channelHeader := &common.ChannelHeader{}
		err := binary.Read(context.Channel, binary.BigEndian, &channelHeader)
		if err != nil {
			break
		}

		if !channelHeader.IsLegal() {
			break
		}

		buf := make([]byte, channelHeader.Len)
		rn, err := io.ReadFull(context.Channel, buf)

		if err != nil || rn != len(buf) {
			break
		}

		session := context.ConnectSessionDict.Find(channelHeader.Id)
		if session == nil {
			log.Printf("not find id [%d] in dict", channelHeader.Id)
			continue
		}

		if channelHeader.IsCmd() && string(buf) == "FIN" {
			finMsg := common.NewMsg(session.Id, nil)
			session.SendQueue.Put(finMsg)
		} else {
			payload := common.NewMsg(channelHeader.Id, buf[rn:])
			session.SendQueue.Put(payload)
		}
	}

	stopMsg := common.NewMsg(0,nil)
	stopMsg.ChannelStop = true
	context.SendQueue.Put(stopMsg)

	context.ConnectSessionDict.FinAll()
}

func doLocalWay(arg []string) {
	pair := strings.Split(arg[2], ":")
	localListenAddr := pair[0] + ":" + pair[1]
	remoteConnectAddr := pair[2] + ":" + pair[3]
	dmuxAddr := arg[3]

	for {
		conn, err := net.Dial("tcp", dmuxAddr)

		if err != nil {
			log.Printf("dial error %v", dmuxAddr)
			time.Sleep(time.Second * 3)
			continue
		}

		connectSynPack := common.NewConnectPack(remoteConnectAddr)
		wn, err := conn.Write(connectSynPack)
		if err != nil || wn != len(connectSynPack) {
			log.Printf("CONNECT error")
			continue
		}

		cmd, err := common.ReadCmdLine(conn)
		if err != nil || !cmd.IsConnectOK() {
			log.Printf("CONNECT error")
			continue
		}

		log.Printf("CONNECT ok %v", dmuxAddr)

		context := common.NewContext(common.Connect, conn)
		context.MultiplexerLocalListener, err = net.Listen("tcp", localListenAddr)
		if err != nil {
			panic(err)
		}

		go serveLocalListener(context)

		serveChannel(context)

		// 在这里关闭,保证重启channel时能listen成功
		context.MultiplexerLocalListener.Close()

		fmt.Printf("read EOF from dmux, close listener: %v", context.MultiplexerLocalListener.Addr())
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
