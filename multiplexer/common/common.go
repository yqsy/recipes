package common

import (
	"net"
	"github.com/yqsy/recipes/blockqueue/blockqueue"
	"sync"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
	"bufio"
	"strconv"
	"log"
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

	Connect = 0
	Bind    = 1
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

func (waterMask *SendWaterMask) DropMask(mask uint32) {
	waterMask.mtx.Lock()
	defer waterMask.mtx.Unlock()
	waterMask.waterMask = 0
	waterMask.cond.Signal()
}

func (waterMask *SendWaterMask) DropMaskTo(mask uint32) {
	waterMask.mtx.Lock()
	defer waterMask.mtx.Unlock()
	waterMask.waterMask = mask
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

	// 专门处理Channel发过来的ACK包,用来降低session的水位
	AckQueue *blockqueue.BlockQueue

	// 接收水位(成功消费掉水位后,向channel发送ack,让对方继续发送数据)
	RecvWaterMask uint32

	// 正确关闭方法:
	// session <=== half close
	// session ===> half close
	CloseCond chan struct{}

	Id uint32

	// 防止channnel本身连接关闭时,因得不到ACK,阻塞在等待水位上
	ChannelIsClose bool
}

func NewSession(conn net.Conn) *Session {
	session := &Session{}
	session.Conn = conn
	session.SendWaterMask = NewSendWaterMask()
	session.SendQueue = blockqueue.NewBlockQueue()
	session.AckQueue = blockqueue.NewBlockQueue()
	session.CloseCond = make(chan struct{}, 2)
	return session
}

type SessionDict struct {
	ConnectSessionDict map[uint32]*Session
	mtx                sync.Mutex
}

func (sessionDict *SessionDict) Append(session *Session) {
	sessionDict.mtx.Lock()
	defer sessionDict.mtx.Unlock()
	sessionDict.ConnectSessionDict[session.Id] = session
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

	for _, session := range sessionDict.ConnectSessionDict {
		finPack := NewFinPack(session.Id)
		session.SendQueue.Put(finPack)
		session.ChannelIsClose = true
		session.SendWaterMask.DropMaskTo(0)
	}
}

func NewSessionDict() *SessionDict {
	sessionDict := &SessionDict{}
	sessionDict.ConnectSessionDict = make(map[uint32]*Session)
	return sessionDict
}

// 作为服务端要同时维护多个context
type Context struct {
	// 读channel阻塞read
	ChannelConn net.Conn

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

	// CONNECT的listener
	MultiplexerLocalListener net.Listener

	// CONNECT所连接的地址
	DmuxConnectAddr string

	// BIND所连接的地址
	MultiplexerConnectAddr string

	// BIND的listener
	DmuxLolcalListener net.Listener

	// 与channel正确关闭方法:
	//  <=== channel half close
	//  ===> channel half close
	CloseCond chan struct{}
}

func NewContext(cmd int, channelConn net.Conn) *Context {
	ctx := &Context{}
	ctx.ChannelConn = channelConn
	ctx.SendQueue = blockqueue.NewBlockQueue()
	ctx.ConnectSessionDict = NewSessionDict()
	ctx.Cmd = cmd
	ctx.IdGen.InitWithMaxId(MaxId)
	ctx.CloseCond = make(chan struct{}, 2)
	return ctx
}

type ChannelPack struct {
	Head ChannelHead
	Body ChannelBody
}

func (channelPack *ChannelPack) Serialize() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, &channelPack.Head)
	buf.Write(channelPack.Body)
	return buf.Bytes()
}

func ReadChannelPack(channelConn net.Conn) (*ChannelPack, error) {
	channelPack := &ChannelPack{}
	err := binary.Read(channelConn, binary.BigEndian, &channelPack.Head)

	if err != nil {
		return nil, err
	}

	if !channelPack.Head.IsLegal() {
		return nil, errors.New(fmt.Sprintf("head isn't legal len:%v id:%v", channelPack.Head.Len, channelPack.Head.Id))
	}

	channelPack.Body = make([]byte, channelPack.Head.Len)
	rn, err := io.ReadFull(channelConn, channelPack.Body)

	if err != nil || rn != len(channelPack.Body) {
		return nil, errors.New("read body error")
	}

	return channelPack, nil
}

func ReadChannelCmd(channelConn net.Conn) (ChannelBody, error) {
	channelPack, err := ReadChannelPack(channelConn)
	if err != nil {
		return ChannelBody{}, err
	} else {
		return channelPack.Body, nil
	}
}

type ChannelHead struct {
	Len uint32
	Id  uint32
	Cmd bool
}

func (channelHeader *ChannelHead) IsLegal() bool {
	if channelHeader.Len > MaxBodyLen || channelHeader.Id > MaxId {
		return false
	} else {
		return true
	}
}

func (channelHeader *ChannelHead) IsCmd() bool {
	if channelHeader.Cmd {
		return true
	} else {
		return false
	}
}

type ChannelBody []byte

func (channelBody ChannelBody) IsConnect() bool {
	if strings.HasPrefix(string(channelBody), "CONNECT") {
		return true
	} else {
		return false
	}
}

func (channelBody ChannelBody) IsConnectOK() bool {
	if string(channelBody) == "CONNECT OK" {
		return true
	} else {
		return false
	}
}

func (channelBody ChannelBody) IsBind() bool {
	if strings.HasPrefix(string(channelBody), "BIND") {
		return true
	} else {
		return false
	}
}

func (channelBody ChannelBody) IsBindOK() bool {
	if string(channelBody) == "BIND OK" {
		return true
	} else {
		return false
	}
}

func (channelBody ChannelBody) IsSyn() bool {
	if string(channelBody) == "SYN" {
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

func (channelBody ChannelBody) IsFin() bool {
	if string(channelBody) == "FIN" {
		return true
	} else {
		return false
	}
}

func (channelBody ChannelBody) IsAck() bool {
	if strings.HasPrefix(string(channelBody), "ACK") {
		return true
	} else {
		return false
	}
}

// TODO: 最好能写成分词语句到结构体的转换
func (channelBody ChannelBody) GetAckBytes() (uint32, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(channelBody)))
	scanner.Split(bufio.ScanWords)

	if !scanner.Scan() || scanner.Text() != "ACK" {
		return 0, errors.New("not ACK")
	}

	if !scanner.Scan() {
		return 0, errors.New("no bytes")
	}

	ackBytes, err := strconv.Atoi(scanner.Text())
	return uint32(ackBytes), err
}

func (channelBody ChannelBody) GetConnectAddr() (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(channelBody)))
	scanner.Split(bufio.ScanWords)

	if !scanner.Scan() || scanner.Text() != "CONNECT" {
		return "", errors.New("not CONNECT")
	}

	if !scanner.Scan() {
		return "", errors.New("no addr")
	}

	connectAddr := scanner.Text()
	return connectAddr, nil
}

func (channelBody ChannelBody) GetBindAddr() (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(channelBody)))
	scanner.Split(bufio.ScanWords)

	if !scanner.Scan() || scanner.Text() != "BIND" {
		return "", errors.New("not BIND")
	}

	if !scanner.Scan() {
		return "", errors.New("no addr")
	}

	bindAddr := scanner.Text()
	return bindAddr, nil
}

func NewPayloadPack(id uint32, body ChannelBody) *ChannelPack {
	channelPack := &ChannelPack{}
	channelPack.Body = body
	channelPack.Head.Len = uint32(len(channelPack.Body))
	channelPack.Head.Id = id
	channelPack.Head.Cmd = false
	return channelPack
}

func NewConnectPack(remoteConnectAddr string) *ChannelPack {
	channelPack := &ChannelPack{}
	channelPack.Body = []byte(fmt.Sprintf("CONNECT %v", remoteConnectAddr))
	channelPack.Head.Len = uint32(len(channelPack.Body))
	channelPack.Head.Id = 0
	channelPack.Head.Cmd = true
	return channelPack
}

func NewConnectAckPack(ok bool) *ChannelPack {
	channelPack := &ChannelPack{}
	if ok {
		channelPack.Body = []byte("CONNECT OK")
	} else {
		channelPack.Body = []byte("CONNECT ERROR")
	}

	channelPack.Head.Len = uint32(len(channelPack.Body))
	channelPack.Head.Id = 0
	channelPack.Head.Cmd = true
	return channelPack
}

func NewBindPack(remoteListenAddr string) *ChannelPack {
	channelPack := &ChannelPack{}
	channelPack.Body = []byte(fmt.Sprintf("BIND %v", remoteListenAddr))
	channelPack.Head.Len = uint32(len(channelPack.Body))
	channelPack.Head.Id = 0
	channelPack.Head.Cmd = true
	return channelPack
}

func NewBindAckPack(ok bool) *ChannelPack {
	channelPack := &ChannelPack{}
	if ok {
		channelPack.Body = []byte("BIND OK")
	} else {
		channelPack.Body = []byte("BIND ERROR")
	}

	channelPack.Head.Len = uint32(len(channelPack.Body))
	channelPack.Head.Id = 0
	channelPack.Head.Cmd = true
	return channelPack
}

func NewSynPack(id uint32) *ChannelPack {
	channelPack := &ChannelPack{}
	channelPack.Body = []byte("SYN")
	channelPack.Head.Len = uint32(len(channelPack.Body))
	channelPack.Head.Id = id
	channelPack.Head.Cmd = true
	return channelPack
}

func NewSynAckPack(id uint32, ok bool) *ChannelPack {
	channelPack := &ChannelPack{}
	if ok {
		channelPack.Body = []byte("SYN OK")
	} else {
		channelPack.Body = []byte("SYN ERROR")
	}

	channelPack.Head.Len = uint32(len(channelPack.Body))
	channelPack.Head.Id = id
	channelPack.Head.Cmd = true
	return channelPack
}

func NewAckPack(id uint32, ackBytes uint32) *ChannelPack {
	channelPack := &ChannelPack{}
	channelPack.Body = []byte(fmt.Sprintf("ACK %v", ackBytes))
	channelPack.Head.Len = uint32(len(channelPack.Body))
	channelPack.Head.Id = id
	channelPack.Head.Cmd = true
	return channelPack
}

func NewFinPack(id uint32) *ChannelPack {
	channelPack := &ChannelPack{}
	channelPack.Body = []byte("FIN")
	channelPack.Head.Len = uint32(len(channelPack.Body))
	channelPack.Head.Id = id
	channelPack.Head.Cmd = true
	return channelPack
}

// -> session
func SessionSendEventLoop(ctx *Context, session *Session) {
	for {
		recvPack := session.SendQueue.Take().(*ChannelPack)

		if recvPack.Head.IsCmd() && recvPack.Body.IsFin() {
			break
		}

		session.Conn.Write(recvPack.Body)
		session.RecvWaterMask += uint32(len(recvPack.Body))
		if session.RecvWaterMask > ResumeWaterMask {
			ackPack := NewAckPack(session.Id, session.RecvWaterMask)
			ctx.SendQueue.Put(ackPack)
			session.RecvWaterMask = 0
		}
	}

	// half close
	session.Conn.(*net.TCPConn).CloseWrite()
	session.CloseCond <- struct{}{}
	log.Printf("[%v]session <- channel done", session.Id)
}

func SessionAckEventLoop(ctx *Context, session *Session) {
	for {
		val := session.AckQueue.Take()

		if val == nil {
			break
		}

		recvPack := val.(*ChannelPack)
		if recvPack.Head.IsCmd() && recvPack.Body.IsAck() {
			ackBytes, err := recvPack.Body.GetAckBytes()
			if err == nil {
				session.SendWaterMask.DropMask(ackBytes)
			}
		}
	}

	log.Printf("[%v]session ack done", session.Id)
}

// <- session
func SessionReadEventLoop(ctx *Context, session *Session) {
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

		payloadPack := NewPayloadPack(session.Id, buf[:rn])
		ctx.SendQueue.Put(payloadPack)
		session.SendWaterMask.RiseMask(uint32(rn))
	}

	// half close
	finPack := NewFinPack(session.Id)
	ctx.SendQueue.Put(finPack)
	session.CloseCond <- struct{}{}
	log.Printf("[%v]session -> channel done", session.Id)

	// stop ACK deal
	session.AckQueue.Put(nil)
}

// -> channel
func ChannelSendEventLoop(ctx *Context) {
	for {

		val := ctx.SendQueue.Take()
		// 退出的接口
		if val == nil {
			break
		}
		sendPack := val.(*ChannelPack)
		sendBytes := sendPack.Serialize()
		wn, err := ctx.ChannelConn.Write(sendBytes)

		if err != nil || wn != len(sendBytes) {
			break
		}
	}

	ctx.CloseCond <- struct{}{}
}

func DispatchToSession(ctx *Context, channelPack *ChannelPack) {
	session := ctx.ConnectSessionDict.Find(channelPack.Head.Id)
	if session == nil {
		var cmd string
		if len(channelPack.Body) < 20 {
			cmd = string(channelPack.Body)
		}
		log.Printf("can't find session id:%v cmd:%v body:%v", channelPack.Head.Id, channelPack.Head.IsCmd(), cmd)

	} else {
		if channelPack.Head.IsCmd() && channelPack.Body.IsAck() {
			session.AckQueue.Put(channelPack)
		} else {
			session.SendQueue.Put(channelPack)
		}
	}
}

// local accept
func ServeSessionActive(ctx *Context, session *Session) {
	defer session.Conn.Close()

	var err error
	session.Id, err = ctx.IdGen.GetFreeId()
	if err != nil {
		log.Printf("session id not enough")
		return
	}
	defer func() {
		ctx.IdGen.ReleaseId(session.Id)
		log.Printf("release id: %v", session.Id)
	}()

	ctx.ConnectSessionDict.Append(session)
	defer ctx.ConnectSessionDict.Del(session.Id)

	// SYN
	synPack := NewSynPack(session.Id)
	ctx.SendQueue.Put(synPack)
	recvPack := session.SendQueue.Take().(*ChannelPack)

	if !recvPack.Head.IsCmd() || !recvPack.Body.IsSynOK() {
		log.Printf("[%v]session SYN error", session.Id)
		return
	}

	log.Printf("[%v]session <-> channel relay", session.Id)

	// session <-block queue channel ack处理
	go SessionAckEventLoop(ctx, session)

	// session <-block queue channel
	go SessionSendEventLoop(ctx, session)

	// session (阻塞read)-> channel
	SessionReadEventLoop(ctx, session)

	// 两边都半关闭完,释放连接
	for i := 0; i < 2; i++ {
		<-session.CloseCond
	}

	log.Printf("[%v]session <-> channel done", session.Id)
}

// local connect
func ServeSessionPassive(ctx *Context, session *Session) {
	defer session.Conn.Close()

	defer ctx.ConnectSessionDict.Del(session.Id)

	log.Printf("[%v]session <-> channel relay", session.Id)

	// session <-block queue channel ack处理
	go SessionAckEventLoop(ctx, session)

	// session <-block queue channel
	go SessionSendEventLoop(ctx, session)

	// session (阻塞read)-> channel
	SessionReadEventLoop(ctx, session)

	// 两边都半关闭完,释放连接
	for i := 0; i < 2; i++ {
		<-session.CloseCond
	}

	log.Printf("[%v]session <-> channel done", session.Id)
}

// 主动SYN方
func ServerChannelActive(ctx *Context) {
	defer ctx.ChannelConn.Close()

	// send channel (block queue , event loop)
	go ChannelSendEventLoop(ctx)

	// channel -> session
	for {
		channelPack, err := ReadChannelPack(ctx.ChannelConn)

		if err != nil {
			break
		}

		DispatchToSession(ctx, channelPack)
	}

	ctx.SendQueue.Put(nil)
	ctx.ConnectSessionDict.FinAll()
	ctx.CloseCond <- struct{}{}

	for i := 0; i < 2; i++ {
		<-ctx.CloseCond
	}
}

// 被动接收SYN,在本地去connect
func ServerChannelPassive(ctx *Context, connectAddr string) {
	// send channel (block queue , event loop)
	go ChannelSendEventLoop(ctx)

	// channel -> session
	for {
		channelPack, err := ReadChannelPack(ctx.ChannelConn)

		if err != nil {
			break
		}

		if channelPack.Head.IsCmd() && channelPack.Body.IsSyn() {
			conn, err := net.Dial("tcp", connectAddr)
			if err != nil {
				log.Printf("dial error %v", err)
				synAckPack := NewSynAckPack(channelPack.Head.Id, false).Serialize()
				ctx.ChannelConn.Write(synAckPack)
			} else {
				synAckPack := NewSynAckPack(channelPack.Head.Id, true).Serialize()
				ctx.ChannelConn.Write(synAckPack)

				session := NewSession(conn)
				session.Id = channelPack.Head.Id
				ctx.ConnectSessionDict.Append(session)
				go ServeSessionPassive(ctx, session)
			}
		} else {
			DispatchToSession(ctx, channelPack)
		}
	}

	ctx.SendQueue.Put(nil)
	ctx.ConnectSessionDict.FinAll()
	ctx.CloseCond <- struct{}{}

	for i := 0; i < 2; i++ {
		<-ctx.CloseCond
	}
}
