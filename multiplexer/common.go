package main

import (
	"net"
	"sync"
	"bytes"
	"encoding/binary"
	"errors"
)

func panicOnError(err error) {
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

func (sessionConn *SessionConn) setConn(conn net.Conn) {
	sessionConn.mtx.Lock()
	defer sessionConn.mtx.Unlock()
	sessionConn.conn = conn
}

func (sessionConn *SessionConn) getConn() net.Conn {
	sessionConn.mtx.Lock()
	defer sessionConn.mtx.Unlock()
	return sessionConn.conn
}

// input or output connections
type MultiConn struct {
	idAndConn map[uint32]net.Conn
	idAndDone map[uint32]chan struct{}
	mtx       sync.Mutex
}

func (multiConn *MultiConn) addConn(id uint32, conn net.Conn) {
	multiConn.mtx.Lock()
	defer multiConn.mtx.Unlock()
	multiConn.idAndConn[id] = conn
	multiConn.idAndDone[id] = make(chan struct{}, 2)
}

func (multiConn *MultiConn) getConn(id uint32) net.Conn {
	multiConn.mtx.Lock()
	defer multiConn.mtx.Unlock()
	if conn, ok := multiConn.idAndConn[id]; ok {
		return conn
	}
	return nil
}

func (multiConn *MultiConn) delConn(id uint32) {
	multiConn.mtx.Lock()
	defer multiConn.mtx.Unlock()
	delete(multiConn.idAndConn, id)
	delete(multiConn.idAndDone, id)
}

func (multiConn *MultiConn) addDone(id uint32) {
	multiConn.mtx.Lock()
	defer multiConn.mtx.Unlock()
	if done, ok := multiConn.idAndDone[id]; ok {
		done <- struct{}{}
	}
}

// input ==> channel ==> output  half close
// input <== channel <== output  half close
// close socket
func (multiConn *MultiConn) waitUntilDie(id uint32) {
	multiConn.mtx.Lock()
	var done = multiConn.idAndDone[id]
	multiConn.mtx.Unlock()
	for i := 0; i < 2; i++ {
		<-done
	}
}

// input <== channel <== output  all connection half close
func (multiConn *MultiConn) shutWriteAllConns() {
	multiConn.mtx.Lock()
	defer multiConn.mtx.Unlock()

	for _, conn := range multiConn.idAndConn {
		conn.(*net.TCPConn).CloseWrite()
	}

	for _, done := range multiConn.idAndDone {
		done <- struct{}{}
	}
}

type IdGen struct {
	ids []uint32
	mtx sync.Mutex
}

func (idGen *IdGen) initWithMaxId(maxId uint32) {
	idGen.mtx.Lock()
	defer idGen.mtx.Unlock()
	for i := uint32(0); i < maxId; i++ {
		idGen.ids = append(idGen.ids, i)
	}
}

func (idGen *IdGen) getFreeId() (uint32, error) {
	idGen.mtx.Lock()
	defer idGen.mtx.Unlock()

	if len(idGen.ids) < 1 {
		return 0, errors.New("no enough ids")
	}

	freeId := idGen.ids[0]
	idGen.ids = idGen.ids[1:]
	return freeId, nil
}

func (idGen *IdGen) releaseFreeId(freeId uint32) {
	idGen.mtx.Lock()
	defer idGen.mtx.Unlock()
	idGen.ids = append(idGen.ids, freeId)
}

func (idGen *IdGen) getFreeIdNum() int {
	idGen.mtx.Lock()
	defer idGen.mtx.Unlock()
	return len(idGen.ids)
}

func generateFinReq(id uint32) []byte {
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

func generatePayload(id uint32, payload []byte) []byte {
	var packetHeader PacketHeader
	packetHeader.Len = uint32(len(payload))
	packetHeader.Id = id
	packetHeader.Cmd = true

	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, &packetHeader)
	buf.Write(payload)
	return buf.Bytes()
}
