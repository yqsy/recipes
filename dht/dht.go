package main

import (
	"crypto/rand"
	"net"
	"github.com/zeebo/bencode"
)

var (
	DhtNodes = []string{
		"router.bittorrent.com:6881",
		"dht.transmissionbt.com:6881",
		"router.utorrent.com:6881"}

	// 	self 20 bytes id
	SelfId = RandomString(20)

	Address = ":6882"
)

type A struct {
	// self id
	Id string `bencode:"id"`

	// target id
	Target string `bencode:"target"`
}

type ReqFindNode struct {
	T string `bencode:"t"`
	Y string `bencode:"y"`
	Q string `bencode:"q"`
	A A      `bencode:"a"`
}

func NewReqFindNode(t, id, target string) *ReqFindNode {
	req := &ReqFindNode{}
	req.T = t
	req.Y = "q"
	req.Q = "find_node"
	req.A.Id = id
	req.A.Target = target
	return req
}

func RandomString(len int) string {
	buf := make([]byte, len)
	rand.Read(buf)
	return string(buf)
}

func main() {
	serverAddr, err := net.ResolveUDPAddr("udp", Address)
	if err != nil {
		panic(err)
	}

	serverConn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		panic(err)
	}

	// send initialization req
	for _, node := range DhtNodes {
		tid := RandomString(2)
		targetId := RandomString(20)
		reqFindNode := NewReqFindNode(tid, SelfId, targetId)

		nodeAddr, err := net.ResolveUDPAddr("udp", node)
		if err != nil {
			panic(err)
		}

		reqBytes, err := bencode.EncodeBytes(reqFindNode)
		if err != nil {
			panic(err)
		}

		wn, err := serverConn.WriteToUDP(reqBytes, nodeAddr)
		_ = wn

		if err != nil {
			panic(err)
		}
	}

}
