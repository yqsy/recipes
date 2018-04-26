package main

import (
	"os"
	"fmt"
	"net"
	"syscall"
	"encoding/binary"
	"log"
	"github.com/yqsy/recipes/socks/socks4"
	"bytes"
	"io"
)

var usage = `Usage:
%v listenAddr socksAddr`

const SO_ORIGINAL_DST = 80

func serve(ctx *Context) {
	defer ctx.localConn.Close()

	file, err := ctx.localConn.(*net.TCPConn).File()
	if err != nil {
		return
	}
	fd := int(file.Fd())

	addr, err := syscall.GetsockoptIPv6Mreq(fd, syscall.IPPROTO_IP, SO_ORIGINAL_DST)
	ctx.originIpv4 = [4]byte{addr.Multiaddr[4], addr.Multiaddr[5],
		addr.Multiaddr[6], addr.Multiaddr[7]}
	ctx.originPort = binary.BigEndian.Uint16([]byte{addr.Multiaddr[2], addr.Multiaddr[3]})
	ctx.originAddr = fmt.Sprintf("%v:%v", ctx.GetIpv4Str(), ctx.originPort)

	log.Printf("originAddr: %v", ctx.originAddr)

	ctx.socks4Conn, err = net.Dial("tcp", ctx.socksAddr)

	if err != nil {
		log.Printf("socks4 server dial error: %v", err)
		return
	}
	defer ctx.socks4Conn.Close()

	socks4Req := socks4.Socks4Req{Version: 0x04, CommandCode: 0x01,
		Port: ctx.originPort, Ipv4Addr: ctx.originIpv4}

	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, &socks4Req)
	// username
	binary.Write(&buf, binary.BigEndian, []byte{0x00})
	wn, err := ctx.socks4Conn.Write(buf.Bytes())
	if err != nil || wn != len(buf.Bytes()) {
		log.Printf("socks4 server handshake error: %v", err)
		return
	}

	var socks4Res socks4.Socks4Res
	err = binary.Read(ctx.socks4Conn, binary.BigEndian, &socks4Res)
	if err != nil {
		log.Printf("socks4 server handshake error: %v", err)
		return
	}

	if !socks4Res.IsSuccess() {
		log.Printf("socks4 server handshake error")
		return
	}

	RelayTcpUntilDie(ctx)
}

func RelayTcpUntilDie(ctx *Context) {
	log.Printf("relay: %v <-> %v\n", ctx.localConn.RemoteAddr(), ctx.socksAddr)

	go func(ctx *Context) {
		io.Copy(ctx.socks4Conn, ctx.localConn)
		ctx.socks4Conn.(*net.TCPConn).CloseWrite()
		log.Printf("done: %v -> %v\n", ctx.localConn.RemoteAddr(), ctx.socksAddr)
		ctx.closeDone <- struct{}{}
	}(ctx)

	go func(ctx *Context) {
		io.Copy(ctx.localConn, ctx.socks4Conn)
		ctx.localConn.(*net.TCPConn).CloseWrite()
		log.Printf("done: %v <- %v\n", ctx.localConn.RemoteAddr(), ctx.socksAddr)
		ctx.closeDone <- struct{}{}
	}(ctx)

	for i := 0; i < 2; i++ {
		<-ctx.closeDone
	}
}

type Context struct {
	socksAddr string

	// 和socks服务的连接
	socks4Conn net.Conn

	// 原始地址
	originPort uint16

	originIpv4 [4]byte

	// 拼接一下地址
	originAddr string

	// 重定向的连接
	localConn net.Conn

	closeDone chan struct{}
}

func (ctx *Context) GetIpv4Str() string {
	return net.IPv4(ctx.originIpv4[0],
		ctx.originIpv4[1],
		ctx.originIpv4[2],
		ctx.originIpv4[3]).String()
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	arg := os.Args

	if len(arg) < 3 {
		fmt.Printf(usage, arg[0])
		return
	}

	listener, err := net.Listen("tcp", arg[1])
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	for {
		context := &Context{socksAddr: arg[2], closeDone: make(chan struct{}, 2)}
		var err error
		context.localConn, err = listener.Accept()
		if err != nil {
			panic(err)
		}
		go serve(context)
	}
}
