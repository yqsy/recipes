package main

import (
	"flag"
	"os"
	"fmt"
	"net"
	"time"
	"encoding/binary"
	"io"
)

func ifErrorExit(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

type SessionMessage struct {
	Number, Length int32
}

func (sessionMessage SessionMessage) print() {
	fmt.Printf("number of buffers = %d\nbuffer length = %d\n",
		sessionMessage.Number, sessionMessage.Length)
	fmt.Printf("%.3f Mib in total\n", sessionMessage.getTotalMb())
}

func (sessionMessage SessionMessage) getTotalMb() float64 {
	totalMb := float64(sessionMessage.Number) * float64(sessionMessage.Length) / 1024.0 / 1024.0
	return totalMb
}

func (sessionMessage SessionMessage) checkLegal() bool {
	if sessionMessage.Length < 0 {
		return false
	} else if sessionMessage.Length > 100*1024*1024 { // 100 MB
		return false
	} else if sessionMessage.Number < 0 {
		return false
	}

	return true
}

func (sessionMessage SessionMessage) checkLegalAndExit() {
	if !sessionMessage.checkLegal() {
		fmt.Println("unlegal SessionMessage")
		os.Exit(-1)
	}
}

func receive(listenAddr string) {
	listener, err := net.Listen("tcp", listenAddr)
	ifErrorExit(err)
	defer listener.Close()
	conn, err := listener.Accept()
	ifErrorExit(err)
	defer conn.Close()

	// read header
	var sessionMessage SessionMessage
	err = binary.Read(conn, binary.BigEndian, &sessionMessage)
	ifErrorExit(err)

	sessionMessage.print()
	sessionMessage.checkLegalAndExit()

	// read payload
	payload := make([]byte, sessionMessage.Length)

	for i := 0; i < int(sessionMessage.Number); i++ {
		// read a payload
		var length int32
		err = binary.Read(conn, binary.BigEndian, &length)
		ifErrorExit(err)
		if length != sessionMessage.Length {
			fmt.Printf("binary.Read %v != %v\n", length, sessionMessage.Length)
			os.Exit(-1)
		}

		n, err := io.ReadFull(conn, payload)
		ifErrorExit(err)
		if n != len(payload) {
			fmt.Printf("io.ReadFull %v != %v\n", n, len(payload))
			os.Exit(-1)
		}

		// write an ack
		err = binary.Write(conn, binary.BigEndian, &length)
		ifErrorExit(err)
	}
}

func transmit(remoteAddr string, number int, length int) {
	sessionMessage := SessionMessage{int32(number), int32(length)}
	sessionMessage.print()
	sessionMessage.checkLegalAndExit()

	conn, err := net.Dial("tcp", remoteAddr)
	ifErrorExit(err)
	defer conn.Close()

	start := time.Now()

	// write header
	err = binary.Write(conn, binary.BigEndian, &sessionMessage)
	ifErrorExit(err)

	// write payload
	payload := make([]byte, 4+length)
	binary.BigEndian.PutUint32(payload, uint32(length))
	for i := 0; i < length; i++ {
		payload[4+i] = "0123456789ABCDEF"[i%16]
	}

	for i := 0; i < number; i++ {
		// write a payload
		n, err := conn.Write(payload)
		ifErrorExit(err)
		if n != len(payload) {
			fmt.Printf("write payload %v != %v\n", n, len(payload))
			os.Exit(-1)
		}

		// read an ack
		var ack int32
		err = binary.Read(conn, binary.BigEndian, &ack)
		ifErrorExit(err)
		if ack != int32(length) {
			fmt.Printf("ack %v != %v\n", ack, int32(length))
			os.Exit(-1)
		}
	}

	elapsed := time.Since(start).Seconds()
	fmt.Printf("%.3f seconds\n%.3f Mib/s\n", elapsed, sessionMessage.getTotalMb()/elapsed)
}

func main() {
	usage := fmt.Sprintf("Usage:\n%v server\n%v client\n", os.Args[0], os.Args[0])

	serverCommand := flag.NewFlagSet("server", flag.ExitOnError)
	serverPort := serverCommand.String("p", "5003", "Listen port")

	clientCommand := flag.NewFlagSet("client", flag.ExitOnError)
	clientHost := clientCommand.String("t", "localhost", "Transmit host")
	clientPort := clientCommand.String("p", "5003", "Transmit port")
	clientNumber := clientCommand.Int("n", 8192, "Number of buffers")
	clientLength := clientCommand.Int("l", 65536, "Buffer length")

	if len(os.Args) < 2 {
		fmt.Printf(usage)
		return
	}

	switch os.Args[1] {
	case "server":
		serverCommand.Parse(os.Args[2:])
	case "client":
		clientCommand.Parse(os.Args[2:])
	default:
		fmt.Printf(usage)
		return
	}

	if serverCommand.Parsed() {
		listenAddr := ":" + *serverPort
		receive(listenAddr)
	}

	if clientCommand.Parsed() {
		remoteAddr := *clientHost + ":" + *clientPort
		transmit(remoteAddr, *clientNumber, *clientLength)
	}
}
