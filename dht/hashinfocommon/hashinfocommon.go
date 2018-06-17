package hashinfocommon

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

type AFindNode struct {
	// self id
	Id string `bencode:"id"`

	// target id
	Target string `bencode:"target"`
}

type ReqFindNode struct {
	T string    `bencode:"t"`
	Y string    `bencode:"y"`
	Q string    `bencode:"q"`
	A AFindNode `bencode:"a"`
}

type RFindNode struct {
	Id string `bencode:"id"`

	// each has id + ip + port
	// 20 bytes + 4 bytes + 2 bytes
	Nodes string `bencode:"nodes"`
}

type ResFindNode struct {
	T string    `bencode:"t"`
	Y string    `bencode:"y"`
	R RFindNode `bencode:"r"`
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

type AGetPeers struct {
	Id       string `bencode:"id"`
	InfoHash string `bencode:"info_hash"`
}

type ReqGetPeers struct {
	T string    `bencode:"t"`
	Y string    `bencode:"y"`
	Q string    `bencode:"q"`
	A AGetPeers `bencode:"a"`
}

type RGetPeers struct {
	Id    string `bencode:"id"`
	Token string `bencode:"token"`
	Nodes string `bencode:"nodes"`
}

type ResGetPeers struct {
	T string    `bencode:"t"`
	Y string    `bencode:"y"`
	R RGetPeers `bencode:"r"`
}

type AAnnouncePeer struct {
	Id          string `bencode:"id"`
	ImpliedPort int    `bencode:"implied_port"`
	InfoHash    string `bencode:"info_hash"`
	Port        int    `bencode:"port"`
	Token       string `bencode:"token"`
}

type ReqAnnouncePeer struct {
	T string        `bencode:"t"`
	Y string        `bencode:"y"`
	Q string        `bencode:"q"`
	A AAnnouncePeer `bencode:"a"`
}

type RAnnouncePeer struct {
	Id string `bencode:"id"`
}

type ResAnnouncePeer struct {
	T string        `bencode:"t"`
	Y string        `bencode:"y"`
	R RAnnouncePeer `bencode:"r"`
}

type APing struct {
	Id string `bencode:"id"`
}

type ReqPing struct {
	T string `bencode:"t"`
	Y string `bencode:"y"`
	Q string `bencode:"q"`
	A APing  `bencode:"a"`
}

type RPing struct {
	Id string `bencode:"id"`
}

type ResPing struct {
	T string `bencode:"t"`
	Y string `bencode:"y"`
	R RPing  `bencode:"r"`
}
