package httpproxy

import (
	"fmt"
	"net"
	"bufio"
	"strings"
	"io"
	"log"
	"net/url"
	"net/textproto"
	"net/http"
)

// parseRequestLine parses "GET /foo HTTP/1.1" into its three parts.
func ParseRequestLine(line string) (method, requestURL, proto string, ok bool) {
	s1 := strings.Index(line, " ")
	s2 := strings.Index(line[s1+1:], " ")
	if s1 < 0 || s2 < 0 {
		return
	}

	s2 += s1 + 1

	return line[:s1], line[s1+1 : s2], line[s2+1:], true
}

func RelayTcpUntilDie(localConn net.Conn, remoteAddr string, remoteConn net.Conn, bufReader *bufio.Reader) {
	log.Printf("relay: %v <-> %v\n", localConn.RemoteAddr(), remoteAddr)
	done := make(chan bool, 2)

	go func(remoteConn net.Conn, localConn net.Conn, remoteAddr string, bufReader *bufio.Reader, done chan bool) {
		io.Copy(remoteConn, bufReader)
		remoteConn.(*net.TCPConn).CloseWrite()
		log.Printf("done: %v -> %v\n", localConn.RemoteAddr(), remoteAddr)
		done <- true
	}(remoteConn, localConn, remoteAddr, bufReader, done)

	go func(localConn net.Conn, remoteConn net.Conn, remoteAddr string, done chan bool) {
		io.Copy(localConn, remoteConn)
		localConn.(*net.TCPConn).CloseWrite()
		log.Printf("done: %v <- %v\n", localConn.RemoteAddr(), remoteAddr)
		done <- true
	}(localConn, remoteConn, remoteAddr, done)

	for i := 0; i < 2; i++ {
		<-done
	}
}

func HttpProxyHandle(localConn net.Conn, bufReader *bufio.Reader) {
	tp := textproto.NewReader(bufReader)

	line, err := tp.ReadLine()
	if err != nil {
		return
	}

	_, requestURL, _, ok := ParseRequestLine(line)

	if !ok {
		return
	}

	URL, err := url.Parse(requestURL)
	if err != nil {
		return
	}

	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		return
	}

	header := http.Header(mimeHeader)

	// support
	// GET /index.html HTTP/1.1
	// Host: www.google.com
	if URL.Host == "" {
		URL.Host = header.Get("Host")
	}

	// qq.com -> qq.com:80
	if strings.Index(URL.Host, ":") < 0 {
		URL.Host = URL.Host + ":80"
	}

	remoteConn, err := net.Dial("tcp", URL.Host)

	if err != nil {
		return
	}

	defer remoteConn.Close()

	remoteConn.Write([]byte(line + "\r\n"))
	header.WriteSubset(remoteConn, nil)
	remoteConn.Write([]byte("\r\n"))
	RelayTcpUntilDie(localConn, URL.Host, remoteConn, bufReader)
}

func HttpsProxyHandle(localConn net.Conn, bufReader *bufio.Reader) {
	tp := textproto.NewReader(bufReader)

	line, err := tp.ReadLine()
	if err != nil {
		return
	}

	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		return
	}
	// do not care
	_ = mimeHeader

	_, requestURL, protocol, ok := ParseRequestLine(line)

	if !ok {
		return
	}

	requestURL = "http://" + requestURL

	URL, err := url.Parse(requestURL)
	if err != nil {
		return
	}

	remoteConn, err := net.Dial("tcp", URL.Host)

	if err != nil {
		return
	}

	defer remoteConn.Close()

	fmt.Fprintf(localConn, "%v 200 Connection established\r\n\r\n", protocol)

	RelayTcpUntilDie(localConn, URL.Host, remoteConn, bufReader)
}
