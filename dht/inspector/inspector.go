package inspector

import (
	"github.com/yqsy/recipes/dht/hashinfocommon"
	"sync"
	"github.com/gin-gonic/gin"
	"net/http"
	"fmt"
)

type Inspector struct {
	Nodes []*hashinfocommon.Node

	// 基础nodes
	BasicNodes []string

	// 自身信息
	SelfId    string
	LocalAddr string

	// 自身发送的find_node请求数
	SendedFindNodeNumber int

	// 自身请求应答,没收到相应应答的事务号
	UnReplyTid map[string]struct{}

	// 收到ping请求数
	ReceivedPingNumber int

	// 收到find_node请求数
	ReceivedFindNodeNumber int

	// 收到get_peers请求数
	ReceivedGetPeersNumber int

	// 收到announce_peer请求数
	ReceivedGetAnnouncePeerNumber int

	// 本次启动到现在收集到的hashinfo
	HashInfoNumberSinceStart int

	// 总计hashinfo
	HashInfoNumberAll int

	mtx sync.Mutex
}

type BasicInfo struct {
	BasicNodes                    []string `json:"BasicNodes"`
	SelfId                        string   `json:"SelfId"`
	LocalAddr                     string   `json:"LocalAddr"`
	NodeNumber                    int      `json:"NodeNumber"`
	SendedFindNodeNumber          int      `json:"SendedFindNodeNumber"`
	UnreplyedNumber               int      `json:"UnreplyedNumber"`
	ReceivedPingNumber            int      `json:"ReceivedPingNumber"`
	ReceivedFindNodeNumber        int      `json:"ReceivedFindNodeNumber"`
	ReceivedGetPeersNumber        int      `json:"ReceivedGetPeersNumber"`
	ReceivedGetAnnouncePeerNumber int      `json:"ReceivedGetAnnouncePeerNumber"`
	HashInfoNumberSinceStart      int      `json:"HashInfoNumberSinceStart"`
	HashInfoNumberAll             int      `json:"HashInfoNumberAll"`
}

type Node struct {
	Id   string `json:"Id"`
	Addr string `json:"Addr"`
}

type AllNodes struct {
	Nodes []Node `json:"Nodes"`
}

func (ins *Inspector) SafeDo(foo func()) {
	ins.mtx.Lock()
	defer ins.mtx.Unlock()
	foo()
}

type HelpInspect struct {
	Ins *Inspector
}

func (help *HelpInspect) BasicInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		basicInfo := &BasicInfo{}

		help.Ins.SafeDo(func() {
			basicInfo.BasicNodes = help.Ins.BasicNodes
			basicInfo.SelfId = help.Ins.SelfId
			basicInfo.LocalAddr = help.Ins.LocalAddr
			basicInfo.NodeNumber = len(help.Ins.Nodes)
			basicInfo.SendedFindNodeNumber = help.Ins.SendedFindNodeNumber
			basicInfo.UnreplyedNumber = len(help.Ins.UnReplyTid)
			basicInfo.ReceivedPingNumber = help.Ins.ReceivedPingNumber
			basicInfo.ReceivedFindNodeNumber = help.Ins.ReceivedFindNodeNumber
			basicInfo.ReceivedGetPeersNumber = help.Ins.ReceivedGetPeersNumber
			basicInfo.ReceivedGetAnnouncePeerNumber = help.Ins.ReceivedGetAnnouncePeerNumber
			basicInfo.HashInfoNumberSinceStart = help.Ins.HashInfoNumberSinceStart
			basicInfo.HashInfoNumberAll = help.Ins.HashInfoNumberAll
		})

		if err := c.BindJSON(basicInfo); err != nil {
			panic(fmt.Sprintf("err: %v", err))
		}

		c.IndentedJSON(http.StatusOK, basicInfo)
	}
}

func (help *HelpInspect) AllNodes() gin.HandlerFunc {
	return func(c *gin.Context) {
		allNodes := &AllNodes{}
		help.Ins.SafeDo(func() {
			for _, node := range help.Ins.Nodes {
				tmpNode := Node{}
				tmpNode.Id = node.Id
				tmpNode.Addr = node.Addr
				allNodes.Nodes = append(allNodes.Nodes, tmpNode)
			}
		})

		if err := c.BindJSON(allNodes); err != nil {
			panic(fmt.Sprintf("err: %v", err))
		}
		c.IndentedJSON(http.StatusOK, allNodes)
	}
}
