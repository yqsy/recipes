package main

import (
	"crypto/tls"
	"net"
	"bufio"
	"fmt"
	"os"
	"log"
	"encoding/gob"
	"github.com/yqsy/recipes/verification/simple_tls_authentication/common"
)

const (
	User     = "admin"
	Password = "123456"
)

func serverLoginUser(conn net.Conn) {
	bufReader := bufio.NewReader(conn)

	line, err := bufReader.ReadString('\n')
	if err != nil {
		log.Println(err)
		return
	}

	log.Print(line)

	wn, err := conn.Write([]byte("world\n"))
	_ = wn
	if err != nil {
		log.Println(err)
		return
	}
}

func safeClose(conn net.Conn) {
	conn.(*tls.Conn).CloseWrite()

	buf := make([]byte, 1024)
	for {
		rn, err := conn.Read(buf)
		_ = rn

		if err != nil {
			break
		}
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	decoder := gob.NewDecoder(conn)
	var loginReq common.LoginReq
	err := decoder.Decode(&loginReq)
	if err != nil {
		log.Printf("err: %v", err)
		return
	}

	if loginReq.User == User && loginReq.Password == Password {
		loginRes := common.LoginRes{true}
		encoder := gob.NewEncoder(conn)
		encoder.Encode(loginRes)
		serverLoginUser(conn)
		safeClose(conn)
	} else {
		loginRes := common.LoginRes{false}
		encoder := gob.NewEncoder(conn)
		encoder.Encode(loginRes)
		safeClose(conn)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	arg := os.Args

	if len(arg) < 2 {
		fmt.Printf("Usage:\n%v listenAddr\n", arg[0])
		return
	}

	cer, err := tls.LoadX509KeyPair("fd.crt", "fd.key")
	if err != nil {
		panic(err)
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}

	listener, err := tls.Listen("tcp", arg[1], config)
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		go handleConn(conn)
	}
}
