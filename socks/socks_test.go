package main

import (
	"testing"
	"net"
	"io"
	"encoding/binary"
	"bytes"
	"bufio"
	"time"
)

// 必须先开启一个socks server监听 :1080

var b = make(chan bool, 1)

func simpleEcho() {
	listener, err := net.Listen("tcp", ":30000")
	if err != nil {
		panic(err)
	}

	defer listener.Close()
	for {
		conn, err := listener.Accept()
		panicOnError(err)

		go func(conn net.Conn) {
			defer conn.Close()
			io.Copy(conn, conn)
		}(conn)
	}
}

func localHalfClose() {
	listener, err := net.Listen("tcp", ":30001")
	if err != nil {
		panic(err)
	}

	defer listener.Close()
	for {
		conn, err := listener.Accept()
		panicOnError(err)

		go func(conn net.Conn) {
			defer conn.Close()
			buffer := make([]byte, 100)
			_, err := conn.Read(buffer)

			if err == nil {
				// err local shut down write first
			}

			_, err = conn.Write([]byte("hello"))
			if err != nil {
				// err
			}

			time.Sleep(time.Second * 1)
			// close, no reset
		}(conn)
	}
}

func remoteHalfClose() {
	listener, err := net.Listen("tcp", ":30002")
	if err != nil {
		panic(err)
	}

	defer listener.Close()
	for {
		conn, err := listener.Accept()
		panicOnError(err)

		go func(conn net.Conn) {
			defer conn.Close()

			conn.(*net.TCPConn).CloseWrite()

			buf := make([]byte, 5)
			n, err := conn.Read(buf)
			if err != nil {
				// err
			}

			if n != len(buf) {
				//err
			}

			if string(buf) == "world" {
				b <- true
			}

		}(conn)
	}
}

func init() {
	go simpleEcho()
	go localHalfClose()
	go remoteHalfClose()
}

func TestSocks4Simple(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1080")
	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()

	socks4Req := Socks4Req{0x04,
		0x01,
		30000,
		[4]byte{127, 0, 0, 1}}

	err = binary.Write(conn, binary.BigEndian, &socks4Req)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// username
	err = binary.Write(conn, binary.BigEndian, []byte{0x00})
	if err != nil {
		t.Fatalf("%v", err)
	}

	var socks4Res Socks4Res
	err = binary.Read(conn, binary.BigEndian, &socks4Res)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if socks4Res.NullByte != 0x00 || socks4Res.States != 0x5A {
		t.Fatalf("NullByte = %v States = %v", socks4Res.NullByte, socks4Res.States)
	}

	testSendAndRecv(conn, t)
}

func TestSocks4ASimple(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1080")
	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()

	socks4Req := Socks4Req{0x04,
		0x01,
		30000,
		[4]byte{0, 0, 0, 255}}

	err = binary.Write(conn, binary.BigEndian, &socks4Req)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// username
	err = binary.Write(conn, binary.BigEndian, []byte{0x00})
	if err != nil {
		t.Fatalf("%v", err)
	}

	err = binary.Write(conn, binary.BigEndian, []byte("localhost"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	// write null
	err = binary.Write(conn, binary.BigEndian, []byte{0x00})
	if err != nil {
		t.Fatalf("%v", err)
	}

	var socks4Res Socks4Res
	err = binary.Read(conn, binary.BigEndian, &socks4Res)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if socks4Res.NullByte != 0x00 || socks4Res.States != 0x5A {
		t.Fatalf("NullByte = %v States = %v", socks4Res.NullByte, socks4Res.States)
	}

	testSendAndRecv(conn, t)
}

func TestHeadFirstBytes(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1080")
	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()

	socks4Req := Socks4Req{0x00, // !
		0x01,
		30000,
		[4]byte{127, 0, 0, 1}}

	var buf bytes.Buffer

	writer := bufio.NewWriter(&buf)

	binary.Write(writer, binary.BigEndian, &socks4Req)
	//binary.Write(writer, binary.BigEndian, []byte{0x00}) // username
	writer.Flush()

	n, err := conn.Write(buf.Bytes())

	if err != nil {
		t.Fatalf("%v", err)
	}

	if n != len(buf.Bytes()) {
		t.Fatalf("conn.Write err")
	}

	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))

	buffer := make([]byte, 100)
	n, err = conn.Read(buffer)
	if err != io.EOF {
		t.Fatalf("remote don't close? %v", err)
	}

	if n != 0 {
		t.Fatalf("return ? value")
	}
}

func TestHeadSecondBytes(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1080")
	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()

	socks4Req := Socks4Req{0x04,
		0x05, // !
		30000,
		[4]byte{127, 0, 0, 1}}

	var buf bytes.Buffer

	writer := bufio.NewWriter(&buf)

	binary.Write(writer, binary.BigEndian, &socks4Req)
	//binary.Write(writer, binary.BigEndian, []byte{0x00}) // username
	writer.Flush()

	n, err := conn.Write(buf.Bytes())

	if err != nil {
		t.Fatalf("%v", err)
	}

	if n != len(buf.Bytes()) {
		t.Fatalf("conn.Write err")
	}

	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))

	buffer := make([]byte, 100)
	n, err = conn.Read(buffer)
	if err != io.EOF {
		t.Fatalf("remote don't close? %v", err)
	}

	if n != 0 {
		t.Fatalf("return ? value")
	}
}

func TestHeadOneByOne(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1080")
	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()

	socks4Req := Socks4Req{0x04,
		0x01,
		30000,
		[4]byte{0, 0, 0, 255}}

	var buf bytes.Buffer

	writer := bufio.NewWriter(&buf)

	binary.Write(writer, binary.BigEndian, &socks4Req)
	binary.Write(writer, binary.BigEndian, []byte{0x00})
	binary.Write(writer, binary.BigEndian, []byte("localhost"))
	binary.Write(writer, binary.BigEndian, []byte{0x00})
	writer.Flush()

	writeBytes := buf.Bytes()[:]
	for i := 0; i < len(writeBytes); i++ {

		n, err := conn.Write([]byte{writeBytes[i]})

		if err != nil {
			t.Fatalf("%v", err)
		}

		if n != 1 {
			t.Fatalf("conn.Write err %v", n)
		}
	}

	var socks4Res Socks4Res
	err = binary.Read(conn, binary.BigEndian, &socks4Res)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if socks4Res.NullByte != 0x00 || socks4Res.States != 0x5A {
		t.Fatalf("NullByte = %v States = %v", socks4Res.NullByte, socks4Res.States)
	}

	testSendAndRecv(conn, t)
}

func TestUserLengthAttack(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1080")
	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()
	socks4Req := Socks4Req{0x04,
		0x01,
		30000,
		[4]byte{0, 0, 0, 255}}

	var buf bytes.Buffer

	writer := bufio.NewWriter(&buf)

	binary.Write(writer, binary.BigEndian, &socks4Req)

	attackBuf := make([]byte, 4096)
	for i := 0; i < len(attackBuf); i++ {
		attackBuf[i] = "123456789ABCDEFG"[i%16]
	}

	writer.Write(attackBuf)
	binary.Write(writer, binary.BigEndian, []byte{0x00})
	binary.Write(writer, binary.BigEndian, []byte("localhost"))
	binary.Write(writer, binary.BigEndian, []byte{0x00})
	writer.Flush()

	n, err := conn.Write(buf.Bytes())

	if err != nil {
		t.Fatalf("%v", err)
	}

	if n != len(buf.Bytes()) {
		t.Fatalf("conn.Write err")
	}

	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))

	buffer := make([]byte, 100)
	n, err = conn.Read(buffer)

	//应该是reset错误, 4096字节正常收到,NULL字节不能正常处理了,缓冲区就4096
	if e, ok := err.(net.Error); ok && e.Timeout() {
		t.Fatalf("remote don't close? %v", err)
	} else if err != nil {
		// This was an error, but not a timeout
	}

	if n != 0 {
		t.Fatalf("return ? value")
	}
}

func TestDomainLengthAttack(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1080")
	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()
	socks4Req := Socks4Req{0x04,
		0x01,
		30000,
		[4]byte{0, 0, 0, 255}}

	var buf bytes.Buffer

	writer := bufio.NewWriter(&buf)

	binary.Write(writer, binary.BigEndian, &socks4Req)

	attackBuf := make([]byte, 4096)
	for i := 0; i < len(attackBuf); i++ {
		attackBuf[i] = "123456789ABCDEFG"[i%16]
	}

	binary.Write(writer, binary.BigEndian, []byte{0x00})
	writer.Write(attackBuf)
	binary.Write(writer, binary.BigEndian, []byte{0x00})
	writer.Flush()

	n, err := conn.Write(buf.Bytes())

	if err != nil {
		t.Fatalf("%v", err)
	}

	if n != len(buf.Bytes()) {
		t.Fatalf("conn.Write err")
	}

	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))

	buffer := make([]byte, 100)
	n, err = conn.Read(buffer)

	//应该是reset错误, 4096字节正常收到,NULL字节不能正常处理了,缓冲区就4096
	if e, ok := err.(net.Error); ok && e.Timeout() {
		t.Fatalf("remote don't close? %v", err)
	} else if err != nil {
		// This was an error, but not a timeout
	}
}

func TestLocalFirstHalfClose(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1080")
	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()
	socks4Req := Socks4Req{0x04,
		0x01,
		30001,
		[4]byte{127, 0, 0, 1}}

	err = binary.Write(conn, binary.BigEndian, &socks4Req)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// username
	err = binary.Write(conn, binary.BigEndian, []byte{0x00})
	if err != nil {
		t.Fatalf("%v", err)
	}

	var socks4Res Socks4Res
	err = binary.Read(conn, binary.BigEndian, &socks4Res)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if socks4Res.NullByte != 0x00 || socks4Res.States != 0x5A {
		t.Fatalf("NullByte = %v States = %v", socks4Res.NullByte, socks4Res.States)
	}

	err = conn.(*net.TCPConn).CloseWrite()
	if err != nil {
		t.Fatalf("%v", err)
	}

	buf := make([]byte, 5)
	n, err := conn.Read(buf)
	// may be closed
	//if err != nil {
	//	t.Fatalf("%v", err)
	//}

	if n != len(buf) {
		t.Fatalf("len = %v %v", n, err)
	}

	if string(buf) != "hello" {
		t.Fatalf("str = %v", string(buf))
	}
}

func TestRemoteFirstHalfClose(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1080")
	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()
	socks4Req := Socks4Req{0x04,
		0x01,
		30002,
		[4]byte{127, 0, 0, 1}}

	err = binary.Write(conn, binary.BigEndian, &socks4Req)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// username
	err = binary.Write(conn, binary.BigEndian, []byte{0x00})
	if err != nil {
		t.Fatalf("%v", err)
	}

	var socks4Res Socks4Res
	err = binary.Read(conn, binary.BigEndian, &socks4Res)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if socks4Res.NullByte != 0x00 || socks4Res.States != 0x5A {
		t.Fatalf("NullByte = %v States = %v", socks4Res.NullByte, socks4Res.States)
	}

	n, err := conn.Write([]byte("world"))

	if err != nil {
		t.Fatalf("%v", err)
	}

	if n != 5 {
		t.Fatalf("%v", n)
	}

	select {
	case <-b:
		// ok
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("remove didn't receive world")
	}

}

func testSendAndRecv(conn net.Conn, t *testing.T) {
	payload := make([]byte, 1024)
	for i := 0; i < len(payload); i++ {
		payload[i] = "0123456789ABCDEF"[i%16]
	}
	n, err := conn.Write(payload)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if n != len(payload) {
		t.Fatalf("conn.Write err", )
	}
	recvPayload := make([]byte, 1024)
	n, err = io.ReadFull(conn, recvPayload)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if n != len(recvPayload) {
		t.Fatalf("conn.Read err", )
	}
	for i := 0; i < len(recvPayload); i++ {
		if recvPayload[i] != "0123456789ABCDEF"[i%16] {
			t.Fatalf("payload err", )
		}
	}
}
