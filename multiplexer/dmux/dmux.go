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
	"io"
	"strconv"
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

	detialConn := globalOutputConns.GetDetialConn(id)
	if detialConn == nil {
		panic("err")
	}

	buf := make([]byte, 32*1024)
	for {
		detialConn.ReadioAndSendChannelControl.WaitCanBeRead()

		rn, err := outputConn.Read(buf)

		if err != nil {
			// send FIN to channel

			finReq := common.GenerateFinReq(id)
			wn, err := sessionConn.Write(finReq)
			if wn != len(finReq) || err != nil {
				globalOutputConns.AddDone(id)
				log.Printf("[%v]force done: (channel)%v <- %v err: %v\n", id, outputConn.LocalAddr(), outputConn.RemoteAddr(), err)
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
			log.Printf("[%v]force done: (channel)%v <- %v err: %v\n", id, outputConn.LocalAddr(), outputConn.RemoteAddr(), err)
			break
		}

		detialConn.ReadioAndSendChannelControl.UpWater(uint32(wn))
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

func waitForChannelAndWriteOutput(detialConn *common.DetialConn, id uint32) {

	sessionConn := globalSessionConn.GetConn()
	if sessionConn == nil {
		// Impossible
		return
	}

	for {
		msg := <-detialConn.SendioQueue

		if msg == nil {
			break
		}

		wn, err := detialConn.Conn.Write(msg.Bytes)
		if err != nil || wn != len(msg.Bytes) {
			break
		}

		detialConn.ReadChannelAndSendioBytes += uint32(len(msg.Bytes))

		if detialConn.ReadChannelAndSendioBytes > common.ConsumeMark {
			ackReq := common.GenerateAckReq(id, detialConn.ReadChannelAndSendioBytes)

			wn, err := sessionConn.Write(ackReq)
			if wn != len(ackReq) || err != nil {
				break
			}

			detialConn.ReadChannelAndSendioBytes = 0
		}
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

	if len(line) < 3 {
		return errors.New("command too short")
	}

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

		detialConn := globalOutputConns.GetDetialConn(packetHeader.Id)
		if detialConn == nil {
			panic("err")
		}

		log.Printf("[%v]relay: (channel)%v <-> %v\n", packetHeader.Id, outputConn.LocalAddr(), outputConn.RemoteAddr())

		go readOutputAndWriteChannel(outputConn, packetHeader.Id)
		go waitForChannelAndWriteOutput(detialConn, packetHeader.Id)

	} else if string(line[:3]) == "FIN" {

		detialConn := globalOutputConns.GetDetialConn(packetHeader.Id)

		if detialConn == nil {
			// may be output not connected

			log.Printf("not find %v\n", packetHeader.Id)
			return nil
		}

		detialConn.Conn.(*net.TCPConn).CloseWrite()
		globalOutputConns.AddDone(packetHeader.Id)
		close(detialConn.SendioQueue)

		log.Printf("[%v]done: (channel)%v -> %v\n", packetHeader.Id, detialConn.Conn.LocalAddr(), detialConn.Conn.RemoteAddr())

	} else if string(line[:3]) == "ACK" {
		detialConn := globalOutputConns.GetDetialConn(packetHeader.Id)

		if detialConn == nil {
			return errors.New("Impossible!")
		}

		if len(line) < 5 {
			return errors.New("command too short")
		}

		ackBytesStr := string(line[4:])
		ackBytes, err := strconv.Atoi(ackBytesStr)

		if err != nil {
			return errors.New("err ack bytes")
		}

		detialConn.ReadioAndSendChannelControl.DownWater(uint32(ackBytes))

		// ACK ok!
		return nil
	} else {
		return errors.New("non supported command")
	}

	return nil
}

func handleChannelPayload(bufReader *bufio.Reader, packetHeader *common.PacketHeader) error {
	buf := make([]byte, packetHeader.Len)
	rn, err := io.ReadFull(bufReader, buf)

	if err != nil {
		return err
	}

	detialConn := globalOutputConns.GetDetialConn(packetHeader.Id)

	if detialConn == nil {
		return nil
	}

	msg := &common.Msg{}
	msg.Bytes = buf[:rn]
	detialConn.SendioQueue <- msg
	return nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	arg := os.Args

	if len(arg) < 2 {
		fmt.Printf("Usage:\n %v listenaddr\nExample:\n %v :30000\n", arg[0], arg[0])
		return
	}

	globalOutputConns = common.NewMultiConn()

	readChannelAndWriteOutput(arg[1])
}
