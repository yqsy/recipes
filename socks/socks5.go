package main

import (
	"net"
	"encoding/binary"
	"log"
	"fmt"
	"bufio"
)

type Socks5GreetingReq struct {
	Version byte // 5

	// 0x00=No authentication
	// 0x01=GSSAPI
	// 0x02=Username/password
	// 0x03–0x7F=methods assigned by IANA
	// 0x80–0xFE=methods reserved for private use
	MethodSupportedNum byte

	// AuthMethod // 1 byte per method supported
}

func (socks5GreetingReq *Socks5GreetingReq) checkLegal(remoteAddr net.Addr) bool {
	if socks5GreetingReq.Version != 0x05 {
		log.Printf("illegal Version: %v from: %v\n",
			socks5GreetingReq.Version, remoteAddr)
		return false
	}

	if socks5GreetingReq.MethodSupportedNum > 5 {
		log.Printf("illegal MethodSupportedNum: %v from: %v\n",
			socks5GreetingReq.MethodSupportedNum, remoteAddr)
		return false
	}
	return true
}

type Socks5GreetingRes struct {
	Version          byte // 5
	ChosenAuthMethod byte // 0xFF=no acceptable methods were offered
}

type Socks5Req struct {
	Version byte // 5

	// 0x01=establish a TCP/IP stream connection
	// 0x02=establish a TCP/IP port binding
	// 0x03=associate a UDP port
	CommandCode byte

	NullByte byte

	// 0x01=IPv4 address
	// 0x03=Domain name
	// 0x04=IPv6 address
	AddressType byte

	// 4 bytes for IPv4 address
	// 1 byte of name length followed by 1–255 bytes the domain name
	// 16 bytes for IPv6 address
	// Destination

	// Port
}

func (socks5Req *Socks5Req) checkLegal(remoteAddr net.Addr) bool {
	if socks5Req.Version != 0x05 {
		log.Printf("illegal Version: %v from: %v\n",
			socks5Req.Version, remoteAddr)
		return false
	}

	if socks5Req.CommandCode != 0x01 {
		log.Printf("illegal CommandCode: %v: from: %v\n", socks5Req.CommandCode, remoteAddr)
		return false
	}

	if ! (socks5Req.AddressType == 0x01 ||
		socks5Req.AddressType == 0x03 ||
		socks5Req.AddressType == 0x04) {
		log.Printf("illegal AddressType: %v: from: %v\n", socks5Req.AddressType, remoteAddr)
		return false
	}

	return true
}

func (socks5Req *Socks5Req) isAddrDomain() bool {
	return socks5Req.AddressType == 0x03
}

type Socks5Res struct {
	Version byte // 5

	//0x00=request granted
	//0x01=general failure
	//0x02=connection not allowed by ruleset
	//0x03=network unreachable
	//0x04=host unreachable
	//0x05=connection refused by destination host
	//0x06=TTL expired
	//0x07=command not supported / protocol error
	//0x08=address type not supported
	Status byte

	NullByte byte

	// 0x01=IPv4 address
	// 0x03=Domain name
	// 0x04=IPv6 address
	AddressType byte

	// 4 bytes for IPv4 address
	// 1 byte of name length followed by 1–255 bytes the domain name
	// 16 bytes for IPv6 address
	ServerBoundAddress [4]byte

	Port uint16
}

func socksHandle5(localConn net.Conn, bufReader *bufio.Reader) {

	var socks5GreetingReq Socks5GreetingReq

	// read 2 bytes greeting header
	err := binary.Read(bufReader, binary.BigEndian, &socks5GreetingReq)
	if err != nil {
		return
	}

	if !socks5GreetingReq.checkLegal(localConn.RemoteAddr()) {
		return
	}

	// read AuthMethods
	AuthMethods := make([]byte, socks5GreetingReq.MethodSupportedNum)

	err = binary.Read(bufReader, binary.BigEndian, AuthMethods)
	if err != nil {
		return
	}

	// no check AuthMethods

	socks5GreetingRes := Socks5GreetingRes{
		0x05,
		0x00}

	// write greeting response
	err = binary.Write(localConn, binary.BigEndian, &socks5GreetingRes)
	if err != nil {
		return
	}

	// read socks5 header
	var socks5Req Socks5Req

	err = binary.Read(bufReader, binary.BigEndian, &socks5Req)
	if err != nil {
		return
	}

	if !socks5Req.checkLegal(localConn.RemoteAddr()) {
		return
	}

	var remoteAddr string

	if socks5Req.isAddrDomain() {
		remoteAddr, err = parseDomainAddr(bufReader)
		if err != nil {
			return
		}

	} else {
		remoteAddr, err = parseIpAddr(&socks5Req, bufReader)
		if err != nil {
			return
		}
	}

	remoteConn, err := net.Dial("tcp", remoteAddr)

	if err != nil {
		return
	}

	defer remoteConn.Close()

	socks5Res := Socks5Res{
		0x05,
		0x00,
		0x00,
		0x01,
		[4]byte{0, 0, 0, 0},
		0}

	err = binary.Write(localConn, binary.BigEndian, &socks5Res)
	if err != nil {
		return
	}
	relayTcpUntilDie(localConn, remoteAddr, remoteConn, bufReader)
}

func parseDomainAddr(bufReader *bufio.Reader) (string, error) {
	var domainLen byte
	err := binary.Read(bufReader, binary.BigEndian, &domainLen)
	if err != nil {
		return "", err
	}

	domainBytes := make([]byte, int(domainLen))
	err = binary.Read(bufReader, binary.BigEndian, &domainBytes)
	if err != nil {
		return "", err
	}

	var port uint16
	err = binary.Read(bufReader, binary.BigEndian, &port)
	if err != nil {
		return "", err
	}

	remoteAddr := string(domainBytes) + ":" + fmt.Sprintf("%v", port)
	return remoteAddr, nil
}

func parseIpAddr(socks5Req *Socks5Req, bufReader *bufio.Reader) (string, error) {

	addressLen := 0
	// IPv4 address
	if socks5Req.AddressType == 0x01 {
		addressLen = 4
	} else if socks5Req.AddressType == 0x04 {
		addressLen = 16
	}

	addr := make([]byte, addressLen)

	err := binary.Read(bufReader, binary.BigEndian, &addr)
	if err != nil {
		return "", err
	}

	var port uint16
	err = binary.Read(bufReader, binary.BigEndian, &port)
	if err != nil {
		return "", err
	}

	remoteAddr := "[" + net.IP(addr).String() + "]" + ":" + fmt.Sprintf("%v", port)
	return remoteAddr, nil
}
