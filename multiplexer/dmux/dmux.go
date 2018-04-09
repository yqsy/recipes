package main

import (
	"os"
	"fmt"
	"net"
	"bufio"
	"encoding/binary"
	"errors"

	"github.com/yqsy/recipes/multiplexer/common"
	"log"
)

var globalSessionConn common.SessionConn

var globalOutputConns *common.MultiConn

// channel <== dmux <== output
func readOutputAndWriteChannel(outputConn net.Conn, id uint32) {
	defer outputConn.Close()

	defer globalOutputConns.DelConn(id)

	sessionConn := globalSessionConn.GetConn()
	if sessionConn == nil {
		// Impossible
		return
	}

	buf := make([]byte, 16384)
	for {
		rn, err := outputConn.Read(buf)

		if err != nil {
			// send FIN to channel

			finReq := common.GenerateFinReq(id)
			wn, err := sessionConn.Write(finReq)
			if wn != len(finReq) || err != nil {
				globalOutputConns.AddDone(id)
				log.Printf("[%v]force done: (channel)%v <- %v\n", id, outputConn.LocalAddr(), outputConn.RemoteAddr())
				break
			}

			globalOutputConns.AddDone(id)
			log.Printf("[%v]done: (channel)%v <- %v\n", id, outputConn.LocalAddr(), outputConn.RemoteAddr())
			break
		}

		// send payload to channel
		payloadReq := common.GeneratePayload(id, buf[:rn])

		wn, err := sessionConn.Write(payloadReq)
		if wn != len(payloadReq) || err != nil {
			globalOutputConns.AddDone(id)
			log.Printf("[%v]force done: (channel)%v <- %v\n", id, outputConn.LocalAddr(), outputConn.RemoteAddr())
			break
		}
	}

	globalOutputConns.WaitUntilDie(id)
}

// channel ==> dmux ==> output
func readChannelAndWriteOutput(listenAddr string) {
	listener, err := net.Listen("tcp", listenAddr)

	common.PanicOnError(err)

	defer listener.Close()

	for {
		sessionConn, err := listener.Accept()

		common.PanicOnError(err)

		globalSessionConn.SetConn(sessionConn)

		log.Printf("session establish\n")

		_ = readChannelAndWriteOutputDetial(sessionConn)

		// err occurred
		globalOutputConns.ShutWriteAllConns(nil, false)

		log.Printf("begin reaccept to %v\n", listenAddr)
	}
}

func readChannelAndWriteOutputDetial(sessionConn net.Conn) error {
	defer sessionConn.Close()

	bufReader := bufio.NewReader(sessionConn)

	for {
		var packetHeader common.PacketHeader

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

	return nil
}

func handleChannelCmd(bufReader *bufio.Reader, packetHeader *common.PacketHeader) error {
	line, err := bufReader.ReadSlice('\n')
	if err != nil {
		return err
	}

	// remove "\r\n"
	if len(line) < 2 {
		return errors.New("command too short")
	}

	line = line[:len(line)-2]

	if string(line[:3]) == "SYN" {
		if len(line) < 5 {
			return errors.New("command too short")
		}

		remoteAddr := line[4:]
		outputConn, err := net.Dial("tcp", string(remoteAddr))
		if err != nil {

			finReq := common.GenerateFinReq(packetHeader.Id)

			sessionConn := globalSessionConn.GetConn()
			if sessionConn == nil {
				// Impossible
				return nil
			}

			wn, err := sessionConn.Write(finReq)
			if wn != len(finReq) || err != nil {
				return nil
			}

			return nil
		}

		globalOutputConns.AddConn(packetHeader.Id, outputConn)

		log.Printf("[%v]relay: (channel)%v <-> %v\n", packetHeader.Id, outputConn.LocalAddr(), outputConn.RemoteAddr())

		go readOutputAndWriteChannel(outputConn, packetHeader.Id)

	} else if string(line[:3]) == "FIN" {

		outPutConn := globalOutputConns.GetConn(packetHeader.Id)

		if outPutConn == nil {
			// may be output not connected
			return nil
		}

		outPutConn.(*net.TCPConn).CloseWrite()
		globalOutputConns.AddDone(packetHeader.Id)

		log.Printf("[%v]done: (channel)%v -> %v\n", packetHeader.Id, outPutConn.LocalAddr(), outPutConn.RemoteAddr())

	} else {
		return errors.New("non supported command")
	}

	return nil
}

func handleChannelPayload(bufReader *bufio.Reader, packetHeader *common.PacketHeader) error {
	buf := make([]byte, 16384)
	rn, err := bufReader.Read(buf)

	if err != nil {
		return err
	}

	go func(packetHeader *common.PacketHeader, buf []byte) {
		outputConn := globalOutputConns.GetConn(packetHeader.Id)

		if outputConn == nil {
			return
		}

		wn, err := outputConn.Write(buf)
		if err != nil || wn != len(buf) {
			return
		}

	}(packetHeader, buf[:rn])

	return nil
}

func main() {
	arg := os.Args

	if len(arg) < 2 {
		fmt.Printf("Usage:\n %v listenaddr\nExample:\n %v :30000\n", arg[0], arg[0])
		return
	}

	globalOutputConns = common.NewMultiConn()

	readChannelAndWriteOutput(arg[1])
}
