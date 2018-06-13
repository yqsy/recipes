package common

import (
	"net"
	"errors"
	"fmt"
	"strconv"
)

// each msg has tid and type
type ProtoSimple struct {
	T string `bencode:"t"`
	Y string `bencode:"y"`
	Q string `bencode:"q"`
}

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

type R struct {
	Id string `bencode:"id"`

	// each has id + ip + port
	// 20 bytes + 4 bytes + 2 bytes
	Nodes string `bencode:"nodes"`
}

type ResFindNode struct {
	T string `bencode:"t"`
	Y string `bencode:"y"`
	R R      `bencode:"r"`
}

type ReqGetPeers struct {

}



func (resFindNode *ResFindNode) CheckValid() error {
	if len(resFindNode.R.Id) != 20 {
		return errors.New(fmt.Sprintf("id len: %v", len(resFindNode.R.Id)))
	} else if len(resFindNode.R.Nodes)%26 != 0 {
		return errors.New(fmt.Sprintf("Nodes len: %v", len(resFindNode.R.Nodes)))
	}
	return nil
}

type Node struct {
	Id   string
	Addr string
}

func (resFindNode *ResFindNode) GetNodes() []Node {
	var nodes []Node

	for i := 0; i < len(resFindNode.R.Nodes)/26; i++ {
		p := i * 26

		if p+25 >= len(resFindNode.R.Nodes) {
			break
		}

		var node Node
		node.Id = resFindNode.R.Nodes[p : p+20]
		p += 20
		ip := net.IPv4(resFindNode.R.Nodes[p],
			resFindNode.R.Nodes[p+1],
			resFindNode.R.Nodes[p+2],
			resFindNode.R.Nodes[p+3]).String()
		p += 4
		port := strconv.Itoa(int(uint16(resFindNode.R.Nodes[p])<<8 | uint16(resFindNode.R.Nodes[p+1])))

		node.Addr = ip + ":" + port

		nodes = append(nodes, node)
	}

	return nodes
}
