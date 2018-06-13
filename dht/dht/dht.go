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

	// prototype pool
	// one req match unique res
	prototypeDict map[string]interface{}

	// for monitor
	Ins inspector.Inspector

	// for join unique check
	uniqueNodePool map[string]struct{}
}

func NewDht() *DHT {
	dht := &DHT{}
	dht.dhtNodes = []string{
		"router.bittorrent.com:6881",
		"dht.transmissionbt.com:6881",
		"router.utorrent.com:6881"}
	dht.selfId = helpful.RandomString(20)
	dht.localAddr = ":6882"
	dht.prototypeDict = make(map[string]interface{})
	dht.uniqueNodePool = make(map[string]struct{})

	dht.Ins.SafeDo(func() {
		dht.Ins.BasicNodes = dht.dhtNodes
		dht.Ins.SelfId = dht.selfId
		dht.Ins.LocalAddr = dht.localAddr
	})

	return dht
}

func (dht *DHT) Serve() error {
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
		reqFindNode := common.NewReqFindNode(tid, dht.selfId, helpful.RandomString(20))

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
		if rn, addr, err := dht.serverConn.ReadFromUDP(buf); err != nil {
			return err
		} else {
			protoSimple := &common.ProtoSimple{}
			if err = bencode.DecodeBytes(buf[:rn], protoSimple); err != nil {
				logrus.Warnf("decode error from: %v", addr)
			} else {
				if protoSimple.Y == "q" {
					dht.HandleReq(buf[:rn], protoSimple.Q)
				} else if protoSimple.Y == "r" {
					dht.HandleRes(buf[:rn], protoSimple.T)
				} else {
					logrus.Warnf(`error "y": %v from: %v`, protoSimple.Y, addr)
				}
			}
		}
	}
}

func (dht *DHT) HandleReq(buf []byte, q string) {
	// dispatch by dict key "q"
	switch q {
	case "ping":
		dht.Ins.SafeDo(func() {
			dht.Ins.ReceivedPingNumber += 1
		})
	case "find_node":
		dht.Ins.SafeDo(func() {
			dht.Ins.ReceivedFindNodeNumber += 1
		})
	case "get_peers":
		dht.Ins.SafeDo(func() {
			dht.Ins.ReceivedGetPeersNumber += 1
		})
	case "announce_peer":
		dht.Ins.SafeDo(func() {
			dht.Ins.ReceivedGetAnnouncePeerNumber += 1
		})
	}
}

func (dht *DHT) HandleRes(buf []byte, tid string) {
	// prototype create res type and dispatch
	if prototype, ok := dht.prototypeDict[tid]; ok {
		delete(dht.prototypeDict, tid)
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
			//selfId := node.Id[:10] + dht.selfId[10:]
			reqFindNode := common.NewReqFindNode(tid, dht.selfId, helpful.RandomString(20))

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
		dht.prototypeDict[tid] = reflect.TypeOf(resType)
		dht.Ins.SafeDo(func() {
			dht.Ins.UnReplyTid[tid] = struct{}{}
			dht.Ins.SendedFindNodeNumber += 1
		})
		return nil
	}
}
