package common

import (
	"net"
	"sync"
	"bytes"
	"encoding/binary"
	"errors"
	"log"
)

func PanicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

type PacketHeader struct {
	Len uint32
	Id  uint32
	Cmd bool

	// body:
	// if Cmd == true
	// command string , end with \r\n
	// else
	// payload
}

// command simple:
// SYN to ip(domain):port\r\n
// FIN\r\n

// ---

// global unique session connection
type SessionConn struct {
	conn net.Conn
	mtx  sync.Mutex
}

func (sessionConn *SessionConn) SetConn(conn net.Conn) {
	sessionConn.mtx.Lock()
	defer sessionConn.mtx.Unlock()
	sessionConn.conn = conn
}

func (sessionConn *SessionConn) GetConn() net.Conn {
	sessionConn.mtx.Lock()
	defer sessionConn.mtx.Unlock()
	return sessionConn.conn
}

// input or output connection
type DetialConn struct {
	conn net.Conn

	// see WaitUntilDie
	done chan struct{}

	// flow control
	readioAndSendChannelBytes uint32
	readChannelAdnSendioBytes uint32
}

func newDetialConn(conn net.Conn) *DetialConn {
	detialConn := &DetialConn{}
	detialConn.conn = conn
	detialConn.done = make(chan struct{}, 2)
	return detialConn
}

// input or output connections
type MultiConn struct {
	connMap map[uint32]*DetialConn
	mtx     sync.Mutex
}

func NewMultiConn() *MultiConn {
	multiConn := &MultiConn{}
	multiConn.connMap = make(map[uint32]*DetialConn)
	return multiConn
}

func (multiConn *MultiConn) AddConn(id uint32, conn net.Conn) {
	multiConn.mtx.Lock()
	defer multiConn.mtx.Unlock()
	multiConn.connMap[id] = newDetialConn(conn)
}

func (multiConn *MultiConn) GetConn(id uint32) net.Conn {
	multiConn.mtx.Lock()
	defer multiConn.mtx.Unlock()
	if detialConn, ok := multiConn.connMap[id]; ok {
		return detialConn.conn
	}
	return nil
}

func (multiConn *MultiConn) getDone(id uint32) chan struct{} {
	multiConn.mtx.Lock()
	defer multiConn.mtx.Unlock()
	if detialConn, ok := multiConn.connMap[id]; ok {
		return detialConn.done
	}
	return nil
}

func (multiConn *MultiConn) DelConn(id uint32) {
	multiConn.mtx.Lock()
	defer multiConn.mtx.Unlock()
	delete(multiConn.connMap, id)
}

func (multiConn *MultiConn) AddDone(id uint32) {
	multiConn.mtx.Lock()
	defer multiConn.mtx.Unlock()
	if detialConn, ok := multiConn.connMap[id]; ok {
		detialConn.done <- struct{}{}
	}
}

// input ==> channel ==> output  half close
// input <== channel <== output  half close
// close socket
func (multiConn *MultiConn) WaitUntilDie(id uint32) {
	done := multiConn.getDone(id)
	if done == nil {
		return
	}
	for i := 0; i < 2; i++ {
		<-done
	}
}

// input [<==] channel <== output  all connection half close
// or
// input ==> channel [==>] output all connection half close
func (multiConn *MultiConn) ShutWriteAllConns(channelConn net.Conn, multiplexer bool) {
	multiConn.mtx.Lock()
	defer multiConn.mtx.Unlock()

	for id, detialConn := range multiConn.connMap {
		detialConn.conn.(*net.TCPConn).CloseWrite()
		detialConn.done <- struct{}{}
		if multiplexer {
			log.Printf("[%v]session force done: %v <- %v(channel)", id, detialConn.conn.RemoteAddr(), channelConn.RemoteAddr())
		} else {
			log.Printf("[%v]session force done: (channel)%v -> %v", id, detialConn.conn.LocalAddr(), detialConn.conn.RemoteAddr())
		}

	}
}

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

func (idGen *IdGen) ReleaseFreeId(freeId uint32) {
	idGen.mtx.Lock()
	defer idGen.mtx.Unlock()
	idGen.ids = append(idGen.ids, freeId)
}

func (idGen *IdGen) GetFreeIdNum() int {
	idGen.mtx.Lock()
	defer idGen.mtx.Unlock()
	return len(idGen.ids)
}

func GenerateFinReq(id uint32) []byte {
	cmd := "FIN\r\n"
	var packetHeader PacketHeader
	packetHeader.Len = uint32(len(cmd))
	packetHeader.Id = id
	packetHeader.Cmd = true

	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, &packetHeader)
	buf.WriteString(cmd)
	return buf.Bytes()
}

func GeneratePayload(id uint32, payload []byte) []byte {
	var packetHeader PacketHeader
	packetHeader.Len = uint32(len(payload))
	packetHeader.Id = id
	packetHeader.Cmd = false

	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, &packetHeader)
	buf.Write(payload)
	return buf.Bytes()
}
