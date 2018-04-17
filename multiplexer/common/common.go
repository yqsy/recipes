package common

import (
	"net"
	"github.com/yqsy/recipes/blockqueue/blockqueue"
	"sync"
	"bytes"
	"encoding/binary"
	"bufio"
	"errors"
	"io"
)

const (
	// 最多让对方每个remoteSession缓存2MiB
	HighWaterMask = 2 * 1024 * 1024

	// 转发了75%后发送ACK让对方下降水位,继续发送
	ResumeWaterMask = HighWaterMask * 0.75
)

const (
	// 包头的Len最多允许4Mib
	MaxPackLen = 4 * 1024 * 1024
)

type SendWaterMask struct {
	waterMask uint32
	mtx       sync.Mutex
	cond      *sync.Cond
}

func NewSendWaterMask() *SendWaterMask {
	sendWaterMask := &SendWaterMask{}
	sendWaterMask.cond = sync.NewCond(&sendWaterMask.mtx)
	return sendWaterMask
}

func (waterMask *SendWaterMask) RiseMask(n uint32) {
	waterMask.mtx.Lock()
	defer waterMask.mtx.Unlock()
	waterMask.waterMask += n
}

func (waterMask *SendWaterMask) WaitUntilCanBeWrite() {
	waterMask.mtx.Lock()

	for !(waterMask.waterMask <= HighWaterMask) {
		waterMask.cond.Wait()
	}
	waterMask.mtx.Unlock()
}

type RemoteSession struct {
	// 读remotesession阻塞read
	// 注意读之前要判断发送水位,水位太高需要等待条件变量
	Session net.Conn

	// 发送水位(向channel发送的水位)
	sendWaterMask *SendWaterMask

	// 写remoteSession阻塞等待blockqueue
	// 写成功后累加接收水位,累加水位到达一定高度时发送ack给channel
	SendQueue *blockqueue.BlockQueue

	// 接收水位(成功消费掉水位后,向channel发送ack,让对方继续发送数据)
	recvWaterMask uint32
}

func NewRemoteSession() *RemoteSession {
	remoteSession := &RemoteSession{}
	remoteSession.sendWaterMask = NewSendWaterMask()
	remoteSession.SendQueue = blockqueue.NewBlockQueue()
	return remoteSession
}

const (
	Connect = 0
	Bind    = 1
)

// 作为服务端要维护多个context,有两种主要的功能1. 类似ssh -NL 2. 类似ssh -NR
type Context struct {
	// 读channel阻塞read
	Channel net.Conn

	// 写channel阻塞等待blockqueue
	SendQueue *blockqueue.BlockQueue

	// [id]session
	ConnectSessionDict map[uint32]*RemoteSession

	// [id]session
	BindSessionDict map[uint32]*RemoteSession

	// connect or bind
	Cmd int
}

func NewContext(cmd int) *Context {
	context := &Context{}
	context.SendQueue = blockqueue.NewBlockQueue()
	context.ConnectSessionDict = make(map[uint32]*RemoteSession)
	context.BindSessionDict = make(map[uint32]*RemoteSession)
	context.Cmd = cmd
	return context
}

// 消息队列传送数据
type Msg struct {
	msg []byte
}

type ChannelPack struct {
	Len uint32
	Id  uint32
	Cmd bool
}

type ChannelPackBytes []byte

func (channelPackBytes ChannelPackBytes) IsConnectOK() bool {
	if string(channelPackBytes) == "CONNECT OK" {
		return true
	} else {
		return false
	}
}

func ReadCmdLine(channelConn net.Conn) (ChannelPackBytes, error) {
	channelPack := &ChannelPack{}

	bufReader := bufio.NewReader(channelConn)
	binary.Read(bufReader, binary.BigEndian, channelPack)

	if channelPack.Len > MaxPackLen {
		return ChannelPackBytes{}, errors.New("packet too long")
	}

	if !channelPack.Cmd {
		return ChannelPackBytes{}, errors.New("not cmd")
	}

	buf := make([]byte, channelPack.Len)
	rn, err := io.ReadFull(bufReader, buf)

	if err != nil || rn != len(buf) {
		return ChannelPackBytes{}, errors.New("cmd error")
	}

	return buf, nil
}

func NewConnectSynPack() ChannelPackBytes {
	pack := []byte("CONNECT")
	channelPack := &ChannelPack{}
	channelPack.Len = uint32(len(pack))
	channelPack.Id = 0
	channelPack.Cmd = true

	var buf bytes.Buffer
	buf.Write(pack)
	binary.Write(&buf, binary.BigEndian, channelPack)
	return buf.Bytes()
}

func NewConnectAckPack() ChannelPackBytes {
	pack := []byte("CONNECT OK")
	channelPack := &ChannelPack{}
	channelPack.Len = uint32(len(pack))
	channelPack.Id = 0
	channelPack.Cmd = true

	var buf bytes.Buffer
	buf.Write(pack)
	binary.Write(&buf, binary.BigEndian, channelPack)
	return buf.Bytes()
}
