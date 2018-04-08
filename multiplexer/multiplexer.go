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
)

var globalSessionConn SessionConn

var globalIdGen IdGen

var globalInputConns MultiConn

func printUsage(exec string) {
	fmt.Printf("Usage:\n"+
		"%v [bind_address]:port:host:hostport remotehost:remoteport\n"+
		"Example:\n"+
		"%v :5001:localhost:5001 pi1:30000\n", exec, exec)
}

// input ==> multiplexer ==> channel
func readInputAndWriteChannel(inputConn net.Conn, remoteConnectAddr string) {
	defer inputConn.Close()

	id, err := globalIdGen.getFreeId()
	if err != nil {
		return
	}
	defer globalIdGen.releaseFreeId(id)

	sessionConn := globalSessionConn.getConn()
	if sessionConn == nil {
		return
	}

	// add conn to global map
	globalInputConns.addConn(id, inputConn)
	defer globalInputConns.delConn(id)

	// send SYN to channel
	synReq := generateSynReq(id, remoteConnectAddr)
	wn, err := sessionConn.Write(synReq)
	if wn != len(synReq) || err != nil {
		return
	}

	buf := make([]byte, 16384)
	for {
		rn, err := inputConn.Read(buf)

		if err != nil {
			// send FIN to channel
			finReq := generateFinReq(id)
			wn, err = sessionConn.Write(finReq)
			if wn != len(finReq) || err != nil {
				return
			}

			globalInputConns.addDone(id)
			break
		}

		// send payload to channel
		payloadReq := generatePayload(id, buf[:rn])

		wn, err = sessionConn.Write(payloadReq)
		if wn != len(payloadReq) || err != nil {
			return
		}
	}

	globalInputConns.waitUntilDie(id)
}

func generateSynReq(id uint32, remoteConnectAddr string) []byte {
	cmd := "SYN to " + remoteConnectAddr + "\r\n"

	var packetHeader PacketHeader
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
		channelConn, err := net.Dial("tcp", remoteAddr)

		if err != nil {
			time.Sleep(time.Second * 1)
			continue
		}

		globalSessionConn.setConn(channelConn)

		_ = readChannelAndWriteInputDetial(channelConn)

		// err occurred
		globalInputConns.shutWriteAllConns()
	}
}

func readChannelAndWriteInputDetial(remoteConn net.Conn) error {
	defer remoteConn.Close()

	bufReader := bufio.NewReader(remoteConn)

	for {
		var packetHeader PacketHeader
		err := binary.Read(bufReader, binary.BigEndian, &packetHeader)
		if err != nil {
			return err
		}

		if packetHeader.Cmd == true {

			err := handleChannelCmd(bufReader, &packetHeader)

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

func handleChannelCmd(bufReader *bufio.Reader, packetHeader *PacketHeader) error {
	line, err := bufReader.ReadSlice('\n')
	if err != nil {
		return err
	}

	// remove "\r"
	if len(line) < 1 {
		return errors.New("command too short")
	}

	line = line[:len(line)-1]

	if string(line) == "FIN" {
		inputConn := globalInputConns.getConn(packetHeader.Id)

		if inputConn == nil {
			return errors.New("Impossible!")
		}

		inputConn.(*net.TCPConn).CloseWrite()
		globalInputConns.addDone(packetHeader.Id)

		// FIN ok!
		return nil
	} else {
		return errors.New("only support FIN command")
	}
}

func handleChannelPayload(bufReader *bufio.Reader, packetHeader *PacketHeader) error {
	buf := make([]byte, 16384)
	rn, err := bufReader.Read(buf)

	if err != nil {
		return err
	}

	go func(packetHeader *PacketHeader, buf []byte) {
		inputConn := globalInputConns.getConn(packetHeader.Id)

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

	// id from 0 ~ 65535
	globalIdGen.initWithMaxId(65536)

	go readChannelAndWriteInput(remoteAddr)

	// accept inputs
	listener, err := net.Listen("tcp", bindAddr)

	panicOnError(err)

	defer listener.Close()

	for {
		inputConn, err := listener.Accept()
		if err != nil {
			continue
		}

		go readInputAndWriteChannel(inputConn, remoteConnectAddr)
	}

}
