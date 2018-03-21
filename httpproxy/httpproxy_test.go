package main

import (
	"net/http"
	"fmt"
	"testing"
	"net"
	"time"
	"bufio"
	"net/textproto"
	"io"
	"strconv"
)

// 必须先开启一个http proxy server 监听:1080

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", r.URL.Path[1:])
}

func init() {
	http.HandleFunc("/", handler)
	go http.ListenAndServe(":30001", nil)
}

func TestHttpSimple(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1080")
	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))

	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()

	fmt.Fprint(conn, "GET http://localhost:30001/hello/world HTTP/1.1\r\n"+
		"User-Agent: curl/7.29.0\r\n"+
		"Host: localhost:30001\r\n"+
		"Accept: */*\r\n"+ "\r\n")

	bufReader := bufio.NewReader(conn)

	tp := textproto.NewReader(bufReader)

	line, err := tp.ReadLine()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if line != "HTTP/1.1 200 OK" {
		t.Fatalf("%v", line)
	}

	mimeHeader, err := tp.ReadMIMEHeader()

	if err != nil {
		t.Fatalf("%v", err)
	}

	header := http.Header(mimeHeader)

	length, err := strconv.Atoi(header["Content-Length"][0])

	buf := make([]byte, length)
	rn, err := io.ReadFull(bufReader, buf)

	if err != nil || rn != length {
		t.Fatal("err")
	}

	if string(buf) != "hello/world" {
		t.Fatal(string(buf))
	}
}

func TestHttpDomainInHeader(t *testing.T) {

	conn, err := net.Dial("tcp", "127.0.0.1:1080")
	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))

	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()

	fmt.Fprint(conn, "GET /hello/world HTTP/1.1\r\n"+
		"User-Agent: curl/7.29.0\r\n"+
		"Host: localhost:30001\r\n"+
		"Accept: */*\r\n"+ "\r\n")

	bufReader := bufio.NewReader(conn)

	tp := textproto.NewReader(bufReader)

	line, err := tp.ReadLine()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if line != "HTTP/1.1 200 OK" {
		t.Fatalf("%v", line)
	}

	mimeHeader, err := tp.ReadMIMEHeader()

	if err != nil {
		t.Fatalf("%v", err)
	}

	header := http.Header(mimeHeader)

	length, err := strconv.Atoi(header["Content-Length"][0])

	buf := make([]byte, length)
	rn, err := io.ReadFull(bufReader, buf)

	if err != nil || rn != length {
		t.Fatal("err")
	}

	if string(buf) != "hello/world" {
		t.Fatal(string(buf))
	}

}

func TestHttpOneByOne(t *testing.T) {

	conn, err := net.Dial("tcp", "127.0.0.1:1080")
	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))

	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()

	req := []byte("GET /hello/world HTTP/1.1\r\n" +
		"User-Agent: curl/7.29.0\r\n" +
		"Host: localhost:30001\r\n" +
		"Accept: */*\r\n" + "\r\n")

	for i := 0; i < len(req); i++ {
		conn.Write([]byte{req[i]})
	}

	bufReader := bufio.NewReader(conn)

	tp := textproto.NewReader(bufReader)

	line, err := tp.ReadLine()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if line != "HTTP/1.1 200 OK" {
		t.Fatalf("%v", line)
	}

	mimeHeader, err := tp.ReadMIMEHeader()

	if err != nil {
		t.Fatalf("%v", err)
	}

	header := http.Header(mimeHeader)

	length, err := strconv.Atoi(header["Content-Length"][0])

	buf := make([]byte, length)
	rn, err := io.ReadFull(bufReader, buf)

	if err != nil || rn != length {
		t.Fatal("err")
	}

	if string(buf) != "hello/world" {
		t.Fatal(string(buf))
	}

}

func TestHttpFirstLineErr(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1080")
	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))

	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()

	fmt.Fprint(conn, "\r\n"+
		"User-Agent: curl/7.29.0\r\n"+
		"Host: localhost:30001\r\n"+
		"Accept: */*\r\n"+ "\r\n")

	buffer := make([]byte, 100)
	rn, err := conn.Read(buffer)

	if rn > 0 {
		t.Fatal(string(rn))
	}
}

func TestHttpNoPort(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1080")
	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))

	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()

	fmt.Fprint(conn, "GET /hello/world HTTP/1.1\r\n"+
		"User-Agent: curl/7.29.0\r\n"+
		"Host: \r\n"+
		"Accept: */*\r\n"+ "\r\n")

	buffer := make([]byte, 100)
	rn, err := conn.Read(buffer)

	if rn > 0 {
		t.Fatal(rn)
	}
}

func TestFirstLineLenAttack(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1080")
	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))

	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()

	fmt.Fprint(conn, "GET http://localhost:30001/hello/world HTTP/1.1")

	buf := make([]byte, 8192)

	conn.Write(buf)

	buffer := make([]byte, 100)
	_, err = conn.Read(buffer)
	if e, ok := err.(net.Error); ok && e.Timeout() {
		t.Fatalf("remote don't close? %v", err)
	} else if err != nil {
		// This was an error, but not a timeout
	}
}
