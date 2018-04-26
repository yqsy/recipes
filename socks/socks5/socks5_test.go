package socks5

import (
	"testing"
	"net"
	"io"
	"encoding/binary"
	"bytes"
	"time"
)

func simpleEcho() {
	listener, err := net.Listen("tcp", ":30003")
	if err != nil {
		panic(err)
	}

	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		go func(conn net.Conn) {
			defer conn.Close()
			io.Copy(conn, conn)
		}(conn)
	}
}

func init() {
	go simpleEcho()
}

func TestSocks5Ipv4(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1080")
	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()

	socks5GreetingReq := []byte{
		0x05,
		0x01,
		0x00}

	wn, err := conn.Write(socks5GreetingReq)
	if err != nil || wn != len(socks5GreetingReq) {
		t.Fatalf("%v", err)
	}

	var socks5GreetingRes Socks5GreetingRes
	err = binary.Read(conn, binary.BigEndian, &socks5GreetingRes)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if socks5GreetingRes.Version != 0x5 || socks5GreetingRes.ChosenAuthMethod != 0x00 {
		t.Fatalf("%v", err)
	}

	socks5Req := [] byte{
		0x05,
		0x01,
		0x00,
		0x01,
		127, 0, 0, 1,
		0x75, 0x33}

	wn, err = conn.Write(socks5Req)
	if err != nil || wn != len(socks5Req) {
		t.Fatalf("%v", err)
	}

	var socks5Res Socks5Res
	err = binary.Read(conn, binary.BigEndian, &socks5Res)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if socks5Res.Version != 5 || socks5Res.Status != 0x00 {
		t.Fatalf("%v", err)
	}

	testSendAndRecv(conn, t)
}

func TestSocks5Domain(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1080")
	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()

	socks5GreetingReq := []byte{
		0x05,
		0x01,
		0x00}

	wn, err := conn.Write(socks5GreetingReq)
	if err != nil || wn != len(socks5GreetingReq) {
		t.Fatalf("%v", err)
	}

	var socks5GreetingRes Socks5GreetingRes
	err = binary.Read(conn, binary.BigEndian, &socks5GreetingRes)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if socks5GreetingRes.Version != 0x5 || socks5GreetingRes.ChosenAuthMethod != 0x00 {
		t.Fatalf("%v", err)
	}

	socks5Req := [] byte{
		0x05,
		0x01,
		0x00,
		0x03,
		0x09, 'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't',
		0x75, 0x33}

	wn, err = conn.Write(socks5Req)
	if err != nil || wn != len(socks5Req) {
		t.Fatalf("%v", err)
	}

	var socks5Res Socks5Res
	err = binary.Read(conn, binary.BigEndian, &socks5Res)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if socks5Res.Version != 5 || socks5Res.Status != 0x00 {
		t.Fatalf("%v", err)
	}

	testSendAndRecv(conn, t)
}

func TestSocks5MultiAuth(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1080")
	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()

	socks5GreetingReq := []byte{
		0x05,
		0x05,
		0x00, 0x01, 0x02, 0x03, 0x80}

	wn, err := conn.Write(socks5GreetingReq)
	if err != nil || wn != len(socks5GreetingReq) {
		t.Fatalf("%v", err)
	}

	var socks5GreetingRes Socks5GreetingRes
	err = binary.Read(conn, binary.BigEndian, &socks5GreetingRes)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if socks5GreetingRes.Version != 0x5 || socks5GreetingRes.ChosenAuthMethod != 0x00 {
		t.Fatalf("%v", err)
	}

	socks5Req := [] byte{
		0x05,
		0x01,
		0x00,
		0x03,
		0x09, 'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't',
		0x75, 0x33}

	wn, err = conn.Write(socks5Req)
	if err != nil || wn != len(socks5Req) {
		t.Fatalf("%v", err)
	}

	var socks5Res Socks5Res
	err = binary.Read(conn, binary.BigEndian, &socks5Res)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if socks5Res.Version != 5 || socks5Res.Status != 0x00 {
		t.Fatalf("%v", err)
	}

	testSendAndRecv(conn, t)
}

func TestSocks5GreetingWithHeader(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1080")
	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()

	socks5GreetingReq := []byte{
		0x05,
		0x01,
		0x00}

	socks5Req := [] byte{
		0x05,
		0x01,
		0x00,
		0x03,
		0x09, 'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't',
		0x75, 0x33}

	var buf bytes.Buffer
	buf.Write(socks5GreetingReq)
	buf.Write(socks5Req)

	wn, err := conn.Write(buf.Bytes())
	if err != nil || wn != len(buf.Bytes()) {
		t.Fatalf("%v", err)
	}

	var socks5GreetingRes Socks5GreetingRes
	err = binary.Read(conn, binary.BigEndian, &socks5GreetingRes)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if socks5GreetingRes.Version != 0x5 || socks5GreetingRes.ChosenAuthMethod != 0x00 {
		t.Fatalf("%v", err)
	}

	var socks5Res Socks5Res
	err = binary.Read(conn, binary.BigEndian, &socks5Res)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if socks5Res.Version != 5 || socks5Res.Status != 0x00 {
		t.Fatalf("%v", err)
	}

	testSendAndRecv(conn, t)
}

func TestSocks5OneByOne(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1080")
	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()

	socks5GreetingReq := []byte{
		0x05,
		0x01,
		0x00}

	socks5Req := [] byte{
		0x05,
		0x01,
		0x00,
		0x03,
		0x09, 'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't',
		0x75, 0x33}

	var buf bytes.Buffer
	buf.Write(socks5GreetingReq)
	buf.Write(socks5Req)

	bytes := buf.Bytes()[:]

	for i := 0; i < len(bytes); i++ {
		wn, err := conn.Write([]byte{bytes[i]})
		if err != nil || wn != 1 {
			t.Fatalf("%v", err)
		}
	}

	var socks5GreetingRes Socks5GreetingRes
	err = binary.Read(conn, binary.BigEndian, &socks5GreetingRes)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if socks5GreetingRes.Version != 0x5 || socks5GreetingRes.ChosenAuthMethod != 0x00 {
		t.Fatalf("%v", err)
	}

	var socks5Res Socks5Res
	err = binary.Read(conn, binary.BigEndian, &socks5Res)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if socks5Res.Version != 5 || socks5Res.Status != 0x00 {
		t.Fatalf("%v", err)
	}

	testSendAndRecv(conn, t)
}

func TestSocks5OverBytes(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1080")
	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()

	socks5GreetingReq := []byte{
		0x05,
		0x01,
		0x00}

	wn, err := conn.Write(socks5GreetingReq)
	if err != nil || wn != len(socks5GreetingReq) {
		t.Fatalf("%v", err)
	}

	var socks5GreetingRes Socks5GreetingRes
	err = binary.Read(conn, binary.BigEndian, &socks5GreetingRes)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if socks5GreetingRes.Version != 0x5 || socks5GreetingRes.ChosenAuthMethod != 0x00 {
		t.Fatalf("%v", err)
	}

	socks5Req := [] byte{
		0x05,
		0x01,
		0x00,
		0x03,
		0x09, 'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't',
		0x75, 0x33, 'h', 'h', 'h'}

	wn, err = conn.Write(socks5Req)
	if err != nil || wn != len(socks5Req) {
		t.Fatalf("%v", err)
	}

	var socks5Res Socks5Res
	err = binary.Read(conn, binary.BigEndian, &socks5Res)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if socks5Res.Version != 5 || socks5Res.Status != 0x00 {
		t.Fatalf("%v", err)
	}

	readBytes := [3]byte{}
	err = binary.Read(conn, binary.BigEndian, &readBytes)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if readBytes != [3]byte{'h', 'h', 'h'} {
		t.Fatalf("%v", err)
	}
}

func TestSocks5VersionErr(t *testing.T) {

	conn, err := net.Dial("tcp", "127.0.0.1:1080")
	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()

	socks5GreetingReq := []byte{
		0x06,
		0x01,
		0x00}

	wn, err := conn.Write(socks5GreetingReq)
	if err != nil || wn != len(socks5GreetingReq) {
		t.Fatalf("%v", err)
	}

	buffer := make([]byte, 100)
	rn, err := conn.Read(buffer)

	if rn > 0 {
		t.Fatal(rn)
	}
}

func TestSocks5CommandCodeErr(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1080")
	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()

	socks5GreetingReq := []byte{
		0x05,
		0x01,
		0x00}

	wn, err := conn.Write(socks5GreetingReq)
	if err != nil || wn != len(socks5GreetingReq) {
		t.Fatalf("%v", err)
	}

	var socks5GreetingRes Socks5GreetingRes
	err = binary.Read(conn, binary.BigEndian, &socks5GreetingRes)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if socks5GreetingRes.Version != 0x5 || socks5GreetingRes.ChosenAuthMethod != 0x00 {
		t.Fatalf("%v", err)
	}

	socks5Req := [] byte{
		0x05,
		0x10,
		0x00,
		0x03,
		0x09, 'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't',
		0x75, 0x33}

	wn, err = conn.Write(socks5Req)
	if err != nil || wn != len(socks5Req) {
		t.Fatalf("%v", err)
	}

	buffer := make([]byte, 100)

	rn, err := conn.Read(buffer)

	if rn > 0 {
		t.Fatal(rn)
	}
}

func TestSocks5AddressTypeErr(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1080")
	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()

	socks5GreetingReq := []byte{
		0x05,
		0x01,
		0x00}

	wn, err := conn.Write(socks5GreetingReq)
	if err != nil || wn != len(socks5GreetingReq) {
		t.Fatalf("%v", err)
	}

	var socks5GreetingRes Socks5GreetingRes
	err = binary.Read(conn, binary.BigEndian, &socks5GreetingRes)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if socks5GreetingRes.Version != 0x5 || socks5GreetingRes.ChosenAuthMethod != 0x00 {
		t.Fatalf("%v", err)
	}

	socks5Req := [] byte{
		0x05,
		0x01,
		0x00,
		0x05,
		0x09, 'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't',
		0x75, 0x33}

	wn, err = conn.Write(socks5Req)
	if err != nil || wn != len(socks5Req) {
		t.Fatalf("%v", err)
	}

	buffer := make([]byte, 100)
	rn, err := conn.Read(buffer)

	if rn > 0 {
		t.Fatal(rn)
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
