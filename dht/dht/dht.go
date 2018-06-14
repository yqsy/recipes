package dht

import (
	"net"
	"github.com/yqsy/recipes/dht/helpful"
	"github.com/yqsy/recipes/dht/transaction"
	"github.com/yqsy/recipes/dht/common"
	"github.com/zeebo/bencode"
	"github.com/Sirupsen/logrus"
	"reflect"
	"github.com/yqsy/recipes/dht/inspector"
)

const (
	TokenLen = 2
)

type DHT struct {
	dhtNodes []string

	// randmon 20 bytes id
	selfId string

	localAddr string

	// send req, get res
	// or get req, reply res
	// all use this udp conn
	serverConn *net.UDPConn

	// gen id
	tm transaction.Transaction

	// res prototype pool
	// one req match unique res
	resPrototypeDict map[string]interface{}

	// req prototype pool
	reqPrototypeDict map[string]interface{}

	// for monitor
	Ins inspector.Inspector

	// for join unique check
	uniqueNodePool map[string]struct{}

	// output
	hashInfoChan chan string
}

func NewDht() *DHT {
	dht := &DHT{}
	dht.dhtNodes = []string{
		"router.bittorrent.com:6881",
		"dht.transmissionbt.com:6881",
		"router.utorrent.com:6881"}
	dht.selfId = helpful.RandomString(20)
	dht.localAddr = ":6882"
	dht.resPrototypeDict = make(map[string]interface{})
	dht.reqPrototypeDict = make(map[string]interface{})
	dht.uniqueNodePool = make(map[string]struct{})
	dht.hashInfoChan = make(chan string, 1024)

	dht.reqPrototypeDict["ping"] = reflect.TypeOf((*common.ReqPing)(nil))
	dht.reqPrototypeDict["find_node"] = reflect.TypeOf((*common.ReqFindNode)(nil))
	dht.reqPrototypeDict["get_peers"] = reflect.TypeOf((*common.ReqGetPeers)(nil))
	dht.reqPrototypeDict["announce_peer"] = reflect.TypeOf((*common.ReqAnnouncePeer)(nil))

	dht.Ins.SafeDo(func() {
		dht.Ins.BasicNodes = dht.dhtNodes
		dht.Ins.SelfId = dht.selfId
		dht.Ins.LocalAddr = dht.localAddr
	})

	return dht
}

func (dht *DHT) Run() error {
	if serverAddr, err := net.ResolveUDPAddr("udp", dht.localAddr); err != nil {
		return err
	} else {
		if dht.serverConn, err = net.ListenUDP("udp", serverAddr); err != nil {
			return err
		} else {
			if err = dht.SendJoin(); err != nil {
				return err
			}
		}
	}

	return dht.RecvAndDispatch()
}

func (dht *DHT) SendJoin() error {
	for _, node := range dht.dhtNodes {
		tid := dht.tm.FetchAndAdd()
		reqFindNode := &common.ReqFindNode{T: tid, Y: "q", Q: "find_node",
			A: common.AFindNode{Id: dht.selfId, Target: helpful.RandomString(20)}}

		if reqBytes, err := bencode.EncodeBytes(reqFindNode); err != nil {
			return err
		} else {
			if nodeAddr, err := net.ResolveUDPAddr("udp", node); err != nil {
				return err
			} else {
				if err = dht.sendReq(reqBytes, nodeAddr, tid, (*common.ResFindNode)(nil)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (dht *DHT) RecvAndDispatch() error {
	buf := make([]byte, 2048)
	for {
		if rn, remoteAddr, err := dht.serverConn.ReadFromUDP(buf); err != nil {
			return err
		} else {
			protoSimple := &common.ProtoSimple{}
			if err = bencode.DecodeBytes(buf[:rn], protoSimple); err != nil {
				logrus.Warnf("decode error from: %v", remoteAddr)
			} else {
				if protoSimple.Y == "q" {
					dht.DispatchReq(buf[:rn], protoSimple.Q, remoteAddr)
				} else if protoSimple.Y == "r" {
					dht.DispatchRes(buf[:rn], protoSimple.T)
				} else {
					logrus.Warnf(`error "y": %v from: %v`, protoSimple.Y, remoteAddr)
				}
			}
		}
	}
}

func (dht *DHT) DispatchReq(buf []byte, q string, remoteAddr *net.UDPAddr) {
	// dispatch by dict key "q"

	if prototype, ok := dht.reqPrototypeDict[q]; ok {

		req := reflect.New(prototype.(reflect.Type).Elem()).Interface()

		if err := bencode.DecodeBytes(buf, req); err != nil {
			logrus.Warnf("decode req err: %v", err)
		} else {
			switch req.(type) {
			case *common.ReqPing:
				dht.Ins.SafeDo(func() {
					dht.Ins.ReceivedPingNumber += 1
				})
			case *common.ReqFindNode:
				dht.Ins.SafeDo(func() {
					dht.Ins.ReceivedFindNodeNumber += 1
				})
			case *common.ReqGetPeers:
				dht.Ins.SafeDo(func() {
					dht.Ins.ReceivedGetPeersNumber += 1
				})

				dht.HandleReqGetPeers(req.(*common.ReqGetPeers), remoteAddr)

			case *common.ReqAnnouncePeer:
				dht.Ins.SafeDo(func() {
					dht.Ins.ReceivedGetAnnouncePeerNumber += 1
				})
			default:
				panic("no way")
			}
		}
	} else {
		logrus.Warnf("can't not find prototype %v", q)
	}
}

func (dht *DHT) HandleReqGetPeers(reqGetPeers *common.ReqGetPeers, remoteAddr *net.UDPAddr) {
	// TODO what is self id?
	// TODO what is token?
	resGetPeers := &common.ResGetPeers{T: reqGetPeers.T, Y: "r",
		R: common.RGetPeers{Id: dht.selfId, Token: reqGetPeers.A.InfoHash[:TokenLen], Nodes: ""}}

	if resBytes, err := bencode.EncodeBytes(resGetPeers); err != nil {
		logrus.Warnf("encode res err: %v", err)
	} else {
		if _, err = dht.serverConn.WriteToUDP(resBytes, remoteAddr); err != nil {
			logrus.Warnf("write udp err: %v", err)
		}
	}
}

func (dht *DHT) HandleReqAnnouncePeer(reqAnnouncePeer *common.ReqAnnouncePeer, remoteAddr *net.UDPAddr) {
	dht.hashInfoChan <- reqAnnouncePeer.A.InfoHash

	// TODO what is self id?
	resAnnouncePeer := &common.ResAnnouncePeer{T: reqAnnouncePeer.T, Y: "r",
		R: common.RAnnouncePeer{Id: dht.selfId}}

	if resBytes, err := bencode.EncodeBytes(resAnnouncePeer); err != nil {
		logrus.Warnf("encode res err: %v", err)
	} else {
		if _, err = dht.serverConn.WriteToUDP(resBytes, remoteAddr); err != nil {
			logrus.Warnf("write udp err: %v", err)
		}
	}
}

func (dht *DHT) DispatchRes(buf []byte, tid string) {
	// prototype create res type and dispatch
	if prototype, ok := dht.resPrototypeDict[tid]; ok {
		delete(dht.resPrototypeDict, tid)
		dht.Ins.SafeDo(func() {
			delete(dht.Ins.UnReplyTid, tid)
		})

		res := reflect.New(prototype.(reflect.Type).Elem()).Interface()
		if err := bencode.DecodeBytes(buf, res); err != nil {
			logrus.Warnf("can't not decode tid: %v err: %v", tid, err)
		} else {
			switch res.(type) {
			case *common.ResFindNode:
				dht.HandleResFindNode(res.(*common.ResFindNode))
			default:
				panic("no way")
			}
		}
	} else {
		logrus.Warnf("not match res received tid: %v,drop it", tid)
	}
}

func (dht *DHT) HandleResFindNode(resFindNode *common.ResFindNode) {
	if err := resFindNode.CheckValid(); err != nil {
		logrus.Warnf("not valid ResFindNode err: %v", err)
	} else {
		nodes := resFindNode.GetNodes()
		logrus.Infof("get %v nodes", len(nodes))
		for _, node := range nodes {
			if _, ok := dht.uniqueNodePool[node.Id]; ok {
				logrus.Infof("node repeat id: %v", node.Id)
				continue
			}
			dht.uniqueNodePool[node.Id] = struct{}{}

			tid := dht.tm.FetchAndAdd()
			// selfId := node.Id[:10] + dht.selfId[10:] TODO what is self id?
			reqFindNode := &common.ReqFindNode{T: tid, Y: "q", Q: "find_node",
				A: common.AFindNode{Id: dht.selfId, Target: helpful.RandomString(20)}}

			if reqBytes, err := bencode.EncodeBytes(reqFindNode); err != nil {
				logrus.Warnf("encode err: %v", err)
			} else {
				if nodeAddr, err := net.ResolveUDPAddr("udp", node.Addr); err != nil {
					logrus.Warnf("resolve udp addr err: %v", err)
				} else {
					if err = dht.sendReq(reqBytes, nodeAddr, tid, (*common.ResFindNode)(nil)); err != nil {
						logrus.Warnf("write udp err: %v", err)
					}
				}
			}
		}
	}
}

func (dht *DHT) sendReq(reqBytes []byte, remoteAddr *net.UDPAddr, tid string, resType interface{}) error {
	if _, err := dht.serverConn.WriteToUDP(reqBytes, remoteAddr); err != nil {
		return err
	} else {
		dht.resPrototypeDict[tid] = reflect.TypeOf(resType)
		dht.Ins.SafeDo(func() {
			dht.Ins.UnReplyTid[tid] = struct{}{}
			dht.Ins.SendedFindNodeNumber += 1
		})
		return nil
	}
}
