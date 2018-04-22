package common

import (
	"net"
	"github.com/yqsy/recipes/blockqueue/blockqueue"
	"sync"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"fmt"
)

const (
	// 最多让对方每个Session缓存2MiB
	HighWaterMask = 2 * 1024 * 1024

	// 转发了75%后发送ACK让对方下降水位,继续发送
	ResumeWaterMask = HighWaterMask * 0.75

	MaxId = 65536
)

const (
	// 包头的Len最多允许4Mib
	MaxBodyLen = 4 * 1024 * 1024
)

type IdGen struct {
	ids []uint32
	mtx sync.Mutex
}

func (idGen *IdGen) InitWithMaxId(maxId uint32) {
	idGen.mtx.Lock()
	defer idGen.mtx.Unlock()
	for i := uint32(0); i < maxId; i++ {
		idGen.ids = append(idGen.ids, i)
	}
}

func (idGen *IdGen) GetFreeId() (uint32, error) {
	idGen.mtx.Lock()
	defer idGen.mtx.Unlock()

	if len(idGen.ids) < 1 {
		return 0, errors.New("no enough ids")
	}

	freeId := idGen.ids[0]
	idGen.ids = idGen.ids[1:]
	return freeId, nil
}

func (idGen *IdGen) ReleaseId(freeId uint32) {
	idGen.mtx.Lock()
	defer idGen.mtx.Unlock()
	idGen.ids = append(idGen.ids, freeId)
}

func (idGen *IdGen) GetFreeIdNum() int {
	idGen.mtx.Lock()
	defer idGen.mtx.Unlock()
	return len(idGen.ids)
}

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

func (waterMask *SendWaterMask) DropMask() {
	waterMask.mtx.Lock()
	defer waterMask.mtx.Unlock()
	waterMask.waterMask = 0
	waterMask.cond.Signal()
}

func (waterMask *SendWaterMask) WaitUntilCanBeWrite() {
	waterMask.mtx.Lock()

	for !(waterMask.waterMask <= HighWaterMask) {
		waterMask.cond.Wait()
	}
	waterMask.mtx.Unlock()
}

type Session struct {
	// 读session阻塞read
	// 注意读之前要判断发送水位,水位太高需要等待条件变量
	Conn net.Conn

	// 发送水位(向channel发送的水位)
	SendWaterMask *SendWaterMask

	// 写Session阻塞等待blockqueue
	// 写成功后累加接收水位,累加水位到达一定高度时发送ack给channel
	SendQueue *blockqueue.BlockQueue

	// 接收水位(成功消费掉水位后,向channel发送ack,让对方继续发送数据)
	RecvWaterMask uint32

	// 正确关闭方法:
	// session <=== half close
	// session ===> half close
	CloseCond chan struct{}

	Id uint32
}

func NewSession(conn net.Conn) *Session {
	session := &Session{}
	session.Conn = conn
	session.SendWaterMask = NewSendWaterMask()
	session.SendQueue = blockqueue.NewBlockQueue()
	session.CloseCond = make(chan struct{}, 2)
	return session
}

const (
	Connect = 0
	Bind    = 1
)

type SessionDict struct {
	ConnectSessionDict map[uint32]*Session
	mtx                sync.Mutex
}

func (sessionDict *SessionDict) Append(id uint32, session *Session) {
	sessionDict.mtx.Lock()
	defer sessionDict.mtx.Unlock()
	sessionDict.ConnectSessionDict[id] = session
}

func (sessionDict *SessionDict) Del(id uint32) {
	sessionDict.mtx.Lock()
	defer sessionDict.mtx.Unlock()
	delete(sessionDict.ConnectSessionDict, id)
}

func (sessionDict *SessionDict) Find(id uint32) *Session {
	sessionDict.mtx.Lock()
	defer sessionDict.mtx.Unlock()
	if val, ok := sessionDict.ConnectSessionDict[id]; ok {
		return val
	} else {
		return nil
	}
}

func (sessionDict *SessionDict) FinAll() {
	sessionDict.mtx.Lock()
	defer sessionDict.mtx.Unlock()

	for _, session := range (sessionDict.ConnectSessionDict) {
		session.SendQueue.Put(nil)
		session.SendWaterMask.DropMask()
	}
}

func NewSessionDict() *SessionDict {
	sessionDict := &SessionDict{}
	sessionDict.ConnectSessionDict = make(map[uint32]*Session)
	return sessionDict
}

// 1. 作为服务端要同时维护多个context
// 2. 有两种主要的功能1. 类似ssh -NL 2. 类似ssh -NR
type Context struct {
	// 读channel阻塞read
	Channel net.Conn

	// 写channel阻塞等待blockqueue
	SendQueue *blockqueue.BlockQueue

	// [id]session
	ConnectSessionDict *SessionDict

	// connect or bind
	Cmd int

	// 序号生成器
	// CONNECT时multiplexer主动生成
	// BIND时dmux主动生成
	IdGen IdGen

	// 作为multiplexer的本地listener
	MultiplexerLocalListener net.Listener


	CloseCond chan struct{}
}

func NewContext(cmd int, channel net.Conn) *Context {
	context := &Context{}
	context.Channel = channel
	context.SendQueue = blockqueue.NewBlockQueue()
	context.ConnectSessionDict = NewSessionDict()
	context.Cmd = cmd
	context.IdGen.InitWithMaxId(MaxId)
	return context
}

// 消息队列传送数据
type Msg struct {
	Id   uint32
	Data []byte

	// 使channel block queue 本身停止工作
	ChannelStop bool
}

func NewMsg(id uint32, data []byte) *Msg {
	msg := &Msg{}
	msg.Id = id
	msg.Data = data
	return msg
}

type ChannelPack []byte

type ChannelHeader struct {
	Len uint32
	Id  uint32
	Cmd bool
}

func (channelHeader *ChannelHeader) IsLegal() bool {
	if channelHeader.Len > MaxBodyLen && channelHeader.Id <= MaxId {
		return false
	} else {
		return true
	}
}

func (channelHeader *ChannelHeader) IsCmd() bool {
	if channelHeader.Cmd {
		return true
	} else {
		return false
	}
}

type ChannelBody []byte

func (channelBody ChannelBody) IsConnectOK() bool {
	if string(channelBody) == "CONNECT OK" {
		return true
	} else {
		return false
	}
}

func (channelBody ChannelBody) IsSynOK() bool {
	if string(channelBody) == "SYN OK" {
		return true
	} else {
		return false
	}
}

func ReadCmdLine(channelConn net.Conn) (ChannelBody, error) {
	channelHeader := &ChannelHeader{}

	binary.Read(channelConn, binary.BigEndian, channelHeader)

	if !channelHeader.IsLegal() {
		return ChannelBody{}, errors.New("packet too long")
	}

	if !channelHeader.IsCmd() {
		return ChannelBody{}, errors.New("not cmd")
	}

	buf := make([]byte, channelHeader.Len)
	rn, err := io.ReadFull(channelConn, buf)

	if err != nil || rn != len(buf) {
		return ChannelBody{}, errors.New("cmd error")
	}

	return buf, nil
}

func NewConnectPack(remoteConnectAddr string) ChannelPack {
	body := []byte(fmt.Sprintf("CONNECT %v", remoteConnectAddr))
	channelHeader := &ChannelHeader{}
	channelHeader.Len = uint32(len(body))
	channelHeader.Id = 0
	channelHeader.Cmd = true

	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, channelHeader)
	buf.Write(body)
	return buf.Bytes()
}

func NewConnectAckPack() ChannelPack {
	body := []byte("CONNECT OK")
	channelHeader := &ChannelHeader{}
	channelHeader.Len = uint32(len(body))
	channelHeader.Id = 0
	channelHeader.Cmd = true

	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, channelHeader)
	buf.Write(body)
	return buf.Bytes()
}
