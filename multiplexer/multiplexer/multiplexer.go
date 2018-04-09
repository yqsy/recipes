package main

import (
	"fmt"
	"net"
	"time"
	"encoding/binary"
	"os"
	"bytes"
	"bufio"
	"errors"
	"strings"
	"log"
	"github.com/yqsy/recipes/multiplexer/common"
)

var globalSessionConn common.SessionConn

var globalIdGen common.IdGen

var globalInputConns *common.MultiConn

func printUsage(exec string) {
	fmt.Printf("Usage:\n"+
		"%v [bind_address]:port:host:hostport remotehost:remoteport\n"+
		"Example:\n"+
		"%v :5001:localhost:5001 pi1:30000\n", exec, exec)
}

// input ==> multiplexer ==> channel
func readInputAndWriteChannel(inputConn net.Conn, remoteConnectAddr string) {
	defer inputConn.Close()

	id, err := globalIdGen.GetFreeId()
	if err != nil {
		return
	}
	defer globalIdGen.ReleaseFreeId(id)

	sessionConn := globalSessionConn.GetConn()
	if sessionConn == nil {
		return
	}

	// add conn to global map
	globalInputConns.AddConn(id, inputConn)
	defer func() {
		globalInputConns.DelConn(id)
		log.Printf("freeid:%v reminder: %v\n", id, globalIdGen.GetFreeIdNum())
	}()

	// send SYN to channel
	synReq := generateSynReq(id, remoteConnectAddr)
	wn, err := sessionConn.Write(synReq)
	if wn != len(synReq) || err != nil {
		return
	}

	log.Printf("[%v]relay: %v <-> %v(channel)\n", id, inputConn.RemoteAddr(), sessionConn.RemoteAddr())

	buf := make([]byte, 16384)
	for {
		rn, err := inputConn.Read(buf)

		if err != nil {
			// send FIN to channel
			finReq := common.GenerateFinReq(id)
			wn, err = sessionConn.Write(finReq)
			if wn != len(finReq) || err != nil {
				globalInputConns.AddDone(id)
				log.Printf("[%v]force done: %v -> %v(channel) err: %v\n", id, inputConn.RemoteAddr(), sessionConn.RemoteAddr(), err)
				break
			}

			globalInputConns.AddDone(id)
			log.Printf("[%v]done: %v -> %v(channel)\n", id, inputConn.RemoteAddr(), sessionConn.RemoteAddr())
			break
		}

		// send payload to channel
		payloadReq := common.GeneratePayload(id, buf[:rn])

		wn, err = sessionConn.Write(payloadReq)
		if wn != len(payloadReq) || err != nil {
			globalInputConns.AddDone(id)
			log.Printf("[%v]force done: %v -> %v(channel) err: %v\n", id, inputConn.RemoteAddr(), sessionConn.RemoteAddr(), err)
			break
		}
	}

	globalInputConns.WaitUntilDie(id)
}

func generateSynReq(id uint32, remoteConnectAddr string) []byte {
	cmd := "SYN " + remoteConnectAddr + "\r\n"

	var packetHeader common.PacketHeader
	packetHeader.Len = uint32(len(cmd))
	packetHeader.Id = id
	packetHeader.Cmd = true

	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, &packetHeader)
	buf.WriteString(cmd)

	return buf.Bytes()
}

// input <== multiplexer <== channel
func readChannelAndWriteInput(remoteAddr string) {
	for {
		sessionConn, err := net.Dial("tcp", remoteAddr)

		if err != nil {
			time.Sleep(time.Second * 1)
			continue
		}

		globalSessionConn.SetConn(sessionConn)

		log.Printf("session establish\n")

		_ = readChannelAndWriteInputDetial(sessionConn)

		// err occurred
		globalInputConns.ShutWriteAllConns(sessionConn, true)

		log.Printf("begin reconnect to %v\n", remoteAddr)
	}
}

func readChannelAndWriteInputDetial(sessionConn net.Conn) error {
	defer sessionConn.Close()

	bufReader := bufio.NewReader(sessionConn)

	for {
		var packetHeader common.PacketHeader
		err := binary.Read(bufReader, binary.BigEndian, &packetHeader)
		if err != nil {
			return err
		}

		if packetHeader.Cmd == true {

			err := handleChannelCmd(bufReader, &packetHeader, sessionConn)

			if err != nil {
				return err
			}

		} else {
			// handle payload

			err := handleChannelPayload(bufReader, &packetHeader)

			if err != nil {
				return err
			}
		}
	}
}

func handleChannelCmd(bufReader *bufio.Reader, packetHeader *common.PacketHeader, sessionConn net.Conn) error {
	line, err := bufReader.ReadSlice('\n')
	if err != nil {
		return err
	}

	// remove "\r\n"
	if len(line) < 2 {
		return errors.New("command too short")
	}

	line = line[:len(line)-2]

	if string(line) == "FIN" {
		inputConn := globalInputConns.GetConn(packetHeader.Id)

		if inputConn == nil {
			return errors.New("Impossible!")
		}

		inputConn.(*net.TCPConn).CloseWrite()
		globalInputConns.AddDone(packetHeader.Id)

		log.Printf("[%v]done: %v <- %v(channel)\n", packetHeader.Id, inputConn.RemoteAddr(), sessionConn.RemoteAddr())

		// FIN ok!
		return nil
	} else {
		return errors.New("only support FIN command")
	}
}

func handleChannelPayload(bufReader *bufio.Reader, packetHeader *common.PacketHeader) error {
	buf := make([]byte, 16384)
	rn, err := bufReader.Read(buf)

	if err != nil {
		return err
	}

	go func(packetHeader *common.PacketHeader, buf []byte) {
		inputConn := globalInputConns.GetConn(packetHeader.Id)

		if inputConn == nil {
			return
		}

		wn, err := inputConn.Write(buf)
		if err != nil || wn != len(buf) {
			return
		}

	}(packetHeader, buf[:rn])

	return nil
}

// :5001:localhost:5001 ==> (:5001, localhost:5001)
func splitArgv(argv1 string) (string, string, error) {
	eles := strings.Split(argv1, ":")
	if len(eles) != 4 {
		return "", "", errors.New("error parameters")
	}
	return eles[0] + ":" + eles[1], eles[2] + ":" + eles[3], nil
}

func main() {
	arg := os.Args

	if len(arg) < 3 {
		printUsage(arg[0])
		return
	}
	// dmux addr
	remoteAddr := arg[2]

	// connection mapping
	bindAddr, remoteConnectAddr, err := splitArgv(arg[1])

	if err != nil {
		printUsage(arg[0])
		return
	}

	globalInputConns = common.NewMultiConn()

	// id from 0 ~ 65535
	globalIdGen.InitWithMaxId(65536)

	go readChannelAndWriteInput(remoteAddr)

	// accept inputs
	listener, err := net.Listen("tcp", bindAddr)

	common.PanicOnError(err)

	defer listener.Close()

	for {
		inputConn, err := listener.Accept()
		if err != nil {
			continue
		}

		go readInputAndWriteChannel(inputConn, remoteConnectAddr)
	}

}
