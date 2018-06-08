package main

import (
	"os"
	"fmt"
	"strconv"
	"net"
	"io"
)

func server(conn *net.UDPConn) {

	buf_ := make([]byte, 2048)

	rn, remoteAddr, err := conn.ReadFromUDP(buf_)
	if err != nil {
		return
	}
	os.Stdout.Write(buf_[:rn])

	fmt.Printf("local:%v remote:%v\n", conn.LocalAddr(), remoteAddr)

	//conn.Close()

	//localAddr, err := net.ResolveUDPAddr("udp", ":5003")
	//if err != nil {
	//	panic(err)
	//}
	//
	//conn2, err := net.DialUDP("udp", localAddr, remoteAddr)
	//
	//if err != nil {
	//	panic(err)
	//}
	//
	//fmt.Printf("remote:%v\n", conn2.RemoteAddr())

	// [stdin] -> remote
	go func() {
		buf := make([]byte, 2048)

		for {
			rn, err := os.Stdin.Read(buf)

			if err != nil {
				break
			}

			conn.WriteTo(buf[:rn], remoteAddr)
		}
	}()

	// stdout <- [remote]
	buf := make([]byte, 2048)
	for {
		rn, remoteAddr, err := conn.ReadFromUDP(buf)

		fmt.Printf("remoteAddr: %v\n", remoteAddr)

		//if remoteAddr.Port != addr.Port || !reflect.DeepEqual(remoteAddr.IP, addr.IP) {
		//	continue
		//}

		if err != nil {
			break
		}
		os.Stdout.Write(buf[:rn])
	}
}

func client(conn net.Conn) {
	defer conn.Close()

	// [stdin] -> remote
	go func() {
		io.Copy(conn, os.Stdin)
	}()

	// stdout <- [remote]
	io.Copy(os.Stdout, conn)
}

func main() {
	arg := os.Args

	if len(arg) < 3 {
		fmt.Printf("Usage:\n  %v -l port\n  %v host port\n", arg[0], arg[0])
		return
	}

	if arg[1] == "-l" {
		// server
		port, err := strconv.Atoi(arg[2])
		if err != nil {
			panic(err)
		}

		addr := net.UDPAddr{Port: port, IP: net.ParseIP("0.0.0.0")}

		conn, err := net.ListenUDP("udp", &addr)
		if err != nil {
			panic(err)
		}

		server(conn)

	} else {
		//client
		addr := arg[1] + ":" + arg[2]
		conn, err := net.Dial("udp", addr)
		if err != nil {
			panic(err)
		}

		client(conn)
	}
}
