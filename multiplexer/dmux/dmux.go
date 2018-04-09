package dmux

import (
	"os"
	"fmt"
	"net"
	"bufio"
	"encoding/binary"
	"errors"

	"github.com/yqsy/recipes/multiplexer/common"
)

var globalSessionConn common.SessionConn

var globalOutputConns common.MultiConn

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
				return
			}

			globalOutputConns.AddDone(id)
			break
		}

		// send payload to channel
		payloadReq := common.GeneratePayload(id, buf[:rn])

		wn, err := sessionConn.Write(payloadReq)
		if wn != len(payloadReq) || err != nil {
			return
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
		channelConn, err := listener.Accept()

		common.PanicOnError(err)

		globalSessionConn.SetConn(channelConn)

		_ = readChannelAndWriteOutputDetial(channelConn)

		// err occurred
		globalOutputConns.ShutWriteAllConns()
	}
}

func readChannelAndWriteOutputDetial(channelConn net.Conn) error {
	defer channelConn.Close()

	bufReader := bufio.NewReader(channelConn)

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

	// remove "\r"
	if len(line) < 1 {
		return errors.New("command too short")
	}

	line = line[:len(line)-1]

	if len(line) < 4 {
		return errors.New("command too short")
	}

	if string(line[:3]) == "SYN" {
		remoteAddr := line[3 : len(line)-3]
		outPutConn, err := net.Dial("tcp", string(remoteAddr))
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

		globalOutputConns.AddConn(packetHeader.Id, outPutConn)
		go readOutputAndWriteChannel(outPutConn, packetHeader.Id)

	} else if string(line[:3]) == "FIN" {

		outPutConn := globalOutputConns.GetConn(packetHeader.Id)

		if outPutConn == nil {
			return errors.New("Impossible!")
		}

		outPutConn.(*net.TCPConn).CloseWrite()
		globalOutputConns.AddDone(packetHeader.Id)

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

	readChannelAndWriteOutput(arg[1])
}
