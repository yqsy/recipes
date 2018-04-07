package main

import (
	"fmt"
	"net"
	"time"
	"sync"
	"github.com/pkg/errors"
	"encoding/binary"
	"os"
	"bytes"
)

var globalSessionConn SessionConn

var globalIdGen IdGen

var globalInputConns MultiConn

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

func printUsage(exec string) {
	fmt.Printf("Usage:\n"+
		"%v [bind_address]:port:host:hostport remotehost:remoteport\n"+
		"Example:\n"+
		"%v :5001:localhost:5001 pi1:30000\n", exec, exec)
}

// input or output connections
type MultiConn struct {
	idAndConn map[uint32]net.Conn
	mtx       sync.Mutex
	done      chan struct{}
}

func (multiConn *MultiConn) addConn(id uint32, conn net.Conn) {
	multiConn.mtx.Lock()
	defer multiConn.mtx.Unlock()
	multiConn.idAndConn[id] = conn
}

func (multiConn *MultiConn) delConn(id uint32) {
	multiConn.mtx.Lock()
	defer multiConn.mtx.Unlock()
	delete(multiConn.idAndConn, id)
}

// input ==> multiplexer ==> channel
func readInputAndWriteChannel(inputConn net.Conn, remoteConnectAddr string) {
	if !writeSynToChannel(remoteConnectAddr) {
		inputConn.Close()
		return
	}

	for {
		buf := make([]byte, 16384)
		_, err := inputConn.Read(buf)

		if err != nil {
			// write eof to channel
		}
	}
}

func writeSynToChannel(remoteConnectAddr string) bool {
	id, err := globalIdGen.getFreeId()
	if err != nil {
		return false
	}
	defer globalIdGen.releaseFreeId(id)

	// send SYN to channel
	synReq := generateSynReq(id, remoteConnectAddr)

	sessionConn := globalSessionConn.getConn()
	if sessionConn == nil {
		return false
	}

	wn, err := sessionConn.Write(synReq)
	if err != nil {
		return false
	}

	if wn != len(synReq) {
		return false
	}
	return true
}

func generateSynReq(id uint32, remoteConnectAddr string) []byte {
	cmd := "SYN to " + remoteConnectAddr + "\r\n"

	var packetHeader PacketHeader
	packetHeader.Len = uint32(len(cmd))
	packetHeader.Id = id
	packetHeader.Cmd = true

	var buf bytes.Buffer
	buf.WriteString(cmd)
	binary.Write(&buf, binary.BigEndian, &packetHeader)

	return buf.Bytes()
}

// input <== multiplexer <== channel
func readChannelAndWriteInput(remoteAddr string) {
	for {
		remoteConn, err := net.Dial("tcp", remoteAddr)

		if err != nil {
			time.Sleep(time.Second * 1)
			continue
		}

		globalSessionConn.setConn(remoteConn)

		_ = readChannelAndWriteInputDetial(remoteConn)
		// err occurred
		// delete all input and reconnect
	}
}

func readChannelAndWriteInputDetial(remoteConn net.Conn) error {
	defer remoteConn.Close()

	for {
		var packetHeader PacketHeader
		err := binary.Read(remoteConn, binary.BigEndian, &packetHeader)
		if err != nil {
			return err
		}

		if packetHeader.Cmd == true {
			handleCmd(remoteConn)
		} else {
			// handle payload
		}
	}
}

func handleCmd(remoteConn net.Conn) {

}

func main() {

	arg := os.Args

	if len(arg) < 3 {
		printUsage(arg[0])
		return
	}

	tmpBindAddr := ":5001"
	tmpRemoteConnectAddr := "localhost:5001"
	tmpRemoteAddr := "localhost:30000"

	// id from 0 ~ 65535
	globalIdGen.initWithMaxId(65536)

	go readChannelAndWriteInput(tmpRemoteAddr)

	// accept inputs
	listener, err := net.Listen("tcp", tmpBindAddr)

	panicOnError(err)

	defer listener.Close()

	for {
		inputConn, err := listener.Accept()
		if err != nil {
			continue
		}

		go readInputAndWriteChannel(inputConn, tmpRemoteConnectAddr)
	}

}
