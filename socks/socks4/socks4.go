package socks4

import (
	"fmt"
	"net"
	"log"
	"encoding/binary"
	"bufio"
	"github.com/yqsy/recipes/socks/common"
)

type Socks4Req struct {
	Version byte // 4

	// 1=TCP/IP stream connection
	// 2=TCP/IP port binding
	CommandCode byte
	Port        uint16
	Ipv4Addr    [4]byte

	// UserId terminated with a null

	// Domain terminated with a null
}

func (socks4Req *Socks4Req) GetIP() string {
	return net.IPv4(socks4Req.Ipv4Addr[0],
		socks4Req.Ipv4Addr[1],
		socks4Req.Ipv4Addr[2],
		socks4Req.Ipv4Addr[3]).String()
}

func (socks4Req *Socks4Req) GetPort() string {
	return fmt.Sprintf("%v", socks4Req.Port)
}

func (socks4Req *Socks4Req) IsSocks4a() bool {
	return socks4Req.Ipv4Addr[0] == 0 &&
		socks4Req.Ipv4Addr[1] == 0 &&
		socks4Req.Ipv4Addr[2] == 0 &&
		socks4Req.Ipv4Addr[3] != 0
}

func (socks4Req *Socks4Req) CheckLegal(remoteAddr net.Addr) bool {
	// must be 0x04 for this version
	if socks4Req.Version != 0x04 {
		log.Printf("illegal Version: %v from: %v\n", socks4Req.Version, remoteAddr)
		return false
	}

	// 0x01 = establish a TCP/IP stream connection
	if socks4Req.CommandCode != 1 {
		log.Printf("illegal CommandCode: %v: from: %v\n", socks4Req.CommandCode, remoteAddr)
		return false
	}
	return true
}

type Socks4Res struct {
	NullByte byte

	// 0x5A=granted 0x5B=rejected or failed
	// 0x5C=client is not running identd
	// 0x5D=client's identd could not confirm the user ID string
	States   byte
	Port     uint16
	Ipv4Addr [4]byte
}

func Socks4Handle(localConn net.Conn, bufReader *bufio.Reader) {

	var socks4Req Socks4Req

	// read 8 bytes header
	err := binary.Read(bufReader, binary.BigEndian, &socks4Req)
	if err != nil {
		return
	}

	if !socks4Req.CheckLegal(localConn.RemoteAddr()) {
		return
	}

	userNameBytes, err := bufReader.ReadSlice(0)

	if err != nil {
		return
	}
	// do not use
	_ = userNameBytes

	var remoteAddr string

	if socks4Req.IsSocks4a() {
		domainBytes, err := bufReader.ReadSlice(0)
		if err != nil {
			return
		}
		remoteAddr = string(domainBytes[:len(domainBytes)-1]) + ":" + socks4Req.GetPort()
	} else {
		remoteAddr = socks4Req.GetIP() + ":" + socks4Req.GetPort()
	}

	remoteConn, err := net.Dial("tcp", remoteAddr)

	if err != nil {
		return
	}

	defer remoteConn.Close()

	// write ack to client
	socks4Res := Socks4Res{
		0x00,
		0x5A,
		0x00,
		[4]byte{0x00, 0x00, 0x00, 0x00}}

	err = binary.Write(localConn, binary.BigEndian, &socks4Res)
	if err != nil {
		return
	}

	common.RelayTcpUntilDie(localConn, remoteAddr, remoteConn, bufReader)
}
