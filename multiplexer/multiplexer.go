package main

import (
	"os"
	"fmt"
	"strings"
	"net"
	"time"
)

// 作为客户端的全局唯一的channel连接
var channelConn *net.Conn = nil

type GolbalConn struct {
	conn *net.Conn
}

func (globalConn *GolbalConn) setConn(conn *net.Conn) {

}

func (globalConn *GolbalConn) getConn() *net.Conn {

}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

type Config struct {
	RemoteAddr string
	ConnPair   map[string]string // [bind addr]remote connect addr
}

func parseConfig(arg []string) *Config {
	if len(arg) < 3 {
		return nil
	}

	config := new(Config)
	config.RemoteAddr = arg[2]

	config.ConnPair = make(map[string]string)

	pairs := strings.Split(arg[1], ";")
	if len(pairs) < 1 {
		return nil
	}
	for _, pair := range pairs {
		fourPart := strings.Split(pair, ":")

		if len(fourPart) != 4 {
			return nil
		}
		bindAddr := fourPart[0] + ":" + fourPart[1]
		remoteConnectAddr := fourPart[2] + ":" + fourPart[3]

		config.ConnPair[bindAddr] = remoteConnectAddr
	}

	return config
}

func printUsage(exec string) {
	fmt.Printf("Usage:\n"+
		"%v [bind_address]:port:host:hostport;[...] remotehost:remoteport\n"+
		"Example:\n"+
		"%v :5001:localhost:5001 pi1:30000\n", exec, exec)
}

// input ==> multiplexer ==> channel
func serveInput(localConn net.Conn, remoteConnectAddr string) {
	defer localConn.Close()

}

func serveChannel(remoteAddr string) {
	// defer remoteConn.Close()

	for {
		remoteConn, err := net.Dial("tcp", remoteAddr)

		if err != nil {
			time.Sleep(time.Second * 1)
			continue
		}

		channelConn = &remoteConn
	}
}

func main() {
	arg := os.Args

	if len(arg) < 3 {
		printUsage(arg[0])
		return
	}

	config := parseConfig(arg)

	if config == nil {
		printUsage(arg[0])
		return
	}

	tmpBindAddr := ":5001"
	tmpRemoteConnectAddr := "localhost:5001"
	tmpRemoteAddr := "localhost:30000"

	// connect channel
	go serveChannel(tmpRemoteAddr)

	// accept inputs
	listener, err := net.Listen("tcp", tmpBindAddr)

	panicOnError(err)

	defer listener.Close()

	for {
		localConn, err := listener.Accept()
		if err != nil {
			continue
		}

		go serveInput(localConn, tmpRemoteConnectAddr)
	}

}
