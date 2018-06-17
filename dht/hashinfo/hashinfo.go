package hashinfo

import (
	"net"
	"github.com/yqsy/recipes/dht/helpful"
	"github.com/yqsy/recipes/dht/transaction"
	"github.com/yqsy/recipes/dht/hashinfocommon"
	"github.com/zeebo/bencode"
	"reflect"
	"github.com/yqsy/recipes/dht/inspector"
	"github.com/yqsy/recipes/dht/metadata"
	"strconv"
	"github.com/op/go-logging"
	"github.com/yqsy/recipes/dht/flowcontrol"
)

var log = logging.MustGetLogger("dht")

const (
	TokenLen = 2
)

type HashInfoGetter struct {
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

	// for join unique check
	uniqueNodePool map[string]struct{}

	// output
	MetaSourceChan chan *metadata.MetaSource

	// for monitor
	Ins *inspector.Inspector

	FlowControl *flowcontrol.FlowControl
}

func NewHashInfoGetter(ins *inspector.Inspector) *HashInfoGetter {
	hg := &HashInfoGetter{}
	hg.dhtNodes = []string{
		"router.bittorrent.com:6881",
		"router.utorrent.com:6881",
		"dht.transmissionbt.com:6881"}
	hg.selfId = helpful.RandomString(20)
	hg.localAddr = ":6881"
	hg.resPrototypeDict = make(map[string]interface{})
	hg.reqPrototypeDict = make(map[string]interface{})
	hg.uniqueNodePool = make(map[string]struct{})
	hg.MetaSourceChan = make(chan *metadata.MetaSource, 1024)
	hg.Ins = ins
	hg.FlowControl = flowcontrol.NewFlowControl()

	hg.reqPrototypeDict["ping"] = reflect.TypeOf((*hashinfocommon.ReqPing)(nil))
	hg.reqPrototypeDict["find_node"] = reflect.TypeOf((*hashinfocommon.ReqFindNode)(nil))
	hg.reqPrototypeDict["get_peers"] = reflect.TypeOf((*hashinfocommon.ReqGetPeers)(nil))
	hg.reqPrototypeDict["announce_peer"] = reflect.TypeOf((*hashinfocommon.ReqAnnouncePeer)(nil))

	hg.Ins.SafeDo(func() {
		hg.Ins.BasicNodes = hg.dhtNodes
		hg.Ins.SelfId = hg.selfId
		hg.Ins.LocalAddr = hg.localAddr
	})

	return hg
}

func (hg *HashInfoGetter) Run() error {
	if serverAddr, err := net.ResolveUDPAddr("udp", hg.localAddr); err != nil {
		return err
	} else {
		if hg.serverConn, err = net.ListenUDP("udp", serverAddr); err != nil {
			return err
		} else {
			if err = hg.SendJoin(); err != nil {
				return err
			}
		}
	}

	go func() {
		hg.FlowControl.Increasing(256)
	}()

	return hg.RecvAndDispatch()
}

func (hg *HashInfoGetter) sendReq(reqBytes []byte, remoteAddr *net.UDPAddr, tid string, resType interface{}) error {
	if _, err := hg.serverConn.WriteToUDP(reqBytes, remoteAddr); err != nil {
		return err
	} else {
		hg.resPrototypeDict[tid] = reflect.TypeOf(resType)
		hg.Ins.SafeDo(func() {
			hg.Ins.UnReplyTid[tid] = struct{}{}
		})
		return nil
	}
}

func (hg *HashInfoGetter) SendJoin() error {
	for _, nodeAddr := range hg.dhtNodes {
		if err := hg.SendFindNode(nodeAddr, hg.selfId, hg.selfId); err != nil {
			return err
		}
	}
	return nil
}

func (hg *HashInfoGetter) SendFindNode(nodeAddr string, selfId, targetId string) error {
	tid := hg.tm.FetchAndAdd()
	reqFindNode := &hashinfocommon.ReqFindNode{T: tid, Y: "q", Q: "find_node",
		A: hashinfocommon.AFindNode{Id: selfId, Target: targetId}}

	if reqBytes, err := bencode.EncodeBytes(reqFindNode); err != nil {
		return err
	} else {
		if nodeAddr, err := net.ResolveUDPAddr("udp", nodeAddr); err != nil {
			return err
		} else {
			if err = hg.sendReq(reqBytes, nodeAddr, tid, (*hashinfocommon.ResFindNode)(nil)); err != nil {
				return err
			}
			hg.Ins.SafeDo(func() {
				hg.Ins.SendedFindNodeNumber += 1
			})
		}
	}
	return nil
}

func (hg *HashInfoGetter) RecvAndDispatch() error {
	buf := make([]byte, 2048)
	for {
		if rn, remoteAddr, err := hg.serverConn.ReadFromUDP(buf); err != nil {
			return err
		} else {
			protoSimple := &hashinfocommon.ProtoSimple{}
			if err = bencode.DecodeBytes(buf[:rn], protoSimple); err != nil {
				log.Warningf("decode error from: %v", remoteAddr)
			} else {
				if protoSimple.Y == "q" {
					hg.DispatchReq(buf[:rn], protoSimple.Q, remoteAddr)
				} else if protoSimple.Y == "r" {
					hg.DispatchRes(buf[:rn], protoSimple.T)
				} else if protoSimple.Y == "e" {
					hg.DispatchError(buf[:rn], protoSimple.T)
				} else {
					log.Warningf("error \"y\": %v from: %v", protoSimple.Y, remoteAddr)
				}
			}
		}
	}
}

func (hg *HashInfoGetter) DispatchReq(buf []byte, q string, remoteAddr *net.UDPAddr) {
	// dispatch by dict key "q"

	if prototype, ok := hg.reqPrototypeDict[q]; ok {

		req := reflect.New(prototype.(reflect.Type).Elem()).Interface()

		if err := bencode.DecodeBytes(buf, req); err != nil {
			log.Warningf("decode req err: %v", err)
		} else {
			switch req.(type) {
			case *hashinfocommon.ReqPing:
				hg.Ins.SafeDo(func() {
					hg.Ins.ReceivedPingNumber += 1
				})
				hg.HandleReqPing(req.(*hashinfocommon.ReqPing), remoteAddr)

			case *hashinfocommon.ReqFindNode:
				hg.Ins.SafeDo(func() {
					hg.Ins.ReceivedFindNodeNumber += 1
				})
				hg.HandleReqFindNode(req.(*hashinfocommon.ReqFindNode), remoteAddr)

			case *hashinfocommon.ReqGetPeers:
				hg.Ins.SafeDo(func() {
					hg.Ins.ReceivedGetPeersNumber += 1
				})
				hg.HandleReqGetPeers(req.(*hashinfocommon.ReqGetPeers), remoteAddr)

			case *hashinfocommon.ReqAnnouncePeer:
				hg.Ins.SafeDo(func() {
					hg.Ins.ReceivedGetAnnouncePeerNumber += 1
				})
				hg.HandleReqAnnouncePeer(req.(*hashinfocommon.ReqAnnouncePeer), remoteAddr)

			default:
				panic("no way")
			}
		}
	} else {
		log.Warningf("can't not find prototype %v", q)
	}
}

func (hg *HashInfoGetter) HandleReqPing(reqPing *hashinfocommon.ReqPing, remoteAddr *net.UDPAddr) {
	resPing := &hashinfocommon.ResPing{T: reqPing.T, Y: "r",
		R: hashinfocommon.RPing{Id: reqPing.A.Id}}

	if resBytes, err := bencode.EncodeBytes(resPing); err != nil {
		log.Warningf("encode res err: %v", err)
	} else {
		if _, err = hg.serverConn.WriteToUDP(resBytes, remoteAddr); err != nil {
			log.Warningf("write udp err: %v", err)
		}
	}
}

func (hg *HashInfoGetter) HandleReqFindNode(reqFineNode *hashinfocommon.ReqFindNode, remoteAddr *net.UDPAddr) {
	resFindNode := &hashinfocommon.ResFindNode{T: reqFineNode.T, Y: "r",
		R: hashinfocommon.RFindNode{Id: hg.selfId, Nodes: ""}}

	if resBytes, err := bencode.EncodeBytes(resFindNode); err != nil {
		log.Warningf("encode res err: %v", err)
	} else {
		if _, err = hg.serverConn.WriteToUDP(resBytes, remoteAddr); err != nil {
			log.Warningf("write udp err: %v", err)
		}
	}
}

func (hg *HashInfoGetter) HandleReqGetPeers(reqGetPeers *hashinfocommon.ReqGetPeers, remoteAddr *net.UDPAddr) {
	// TODO what is self id?
	// TODO what is token?
	resGetPeers := &hashinfocommon.ResGetPeers{T: reqGetPeers.T, Y: "r",
		R: hashinfocommon.RGetPeers{Id: hg.selfId, Token: reqGetPeers.A.InfoHash[:TokenLen], Nodes: ""}}

	if resBytes, err := bencode.EncodeBytes(resGetPeers); err != nil {
		log.Warningf("encode res err: %v", err)
	} else {
		if _, err = hg.serverConn.WriteToUDP(resBytes, remoteAddr); err != nil {
			log.Warningf("write udp err: %v", err)
		}
	}
}

func (hg *HashInfoGetter) HandleReqAnnouncePeer(reqAnnouncePeer *hashinfocommon.ReqAnnouncePeer, remoteAddr *net.UDPAddr) {
	if len(reqAnnouncePeer.A.InfoHash) != 20 {
		log.Warningf("infohash len != 20")
		return
	}

	peerAddr := remoteAddr.IP.String() + ":" + strconv.Itoa(reqAnnouncePeer.A.Port)
	hg.MetaSourceChan <- &metadata.MetaSource{
		Hashinfo: reqAnnouncePeer.A.InfoHash,
		Addr:     peerAddr}

	// TODO what is self id?
	resAnnouncePeer := &hashinfocommon.ResAnnouncePeer{T: reqAnnouncePeer.T, Y: "r",
		R: hashinfocommon.RAnnouncePeer{Id: hg.selfId}}

	if resBytes, err := bencode.EncodeBytes(resAnnouncePeer); err != nil {
		log.Warningf("encode res err: %v", err)
	} else {
		if _, err = hg.serverConn.WriteToUDP(resBytes, remoteAddr); err != nil {
			log.Warningf("write udp err: %v", err)
		}
	}
}

func (hg *HashInfoGetter) DispatchRes(buf []byte, tid string) {
	// prototype create res type and dispatch
	if prototype, ok := hg.resPrototypeDict[tid]; ok {
		delete(hg.resPrototypeDict, tid)
		hg.Ins.SafeDo(func() {
			delete(hg.Ins.UnReplyTid, tid)
		})

		res := reflect.New(prototype.(reflect.Type).Elem()).Interface()
		if err := bencode.DecodeBytes(buf, res); err != nil {
			log.Warningf("can't not decode tid: %v err: %v", helpful.Get10Hex(tid), err)
		} else {
			switch res.(type) {
			case *hashinfocommon.ResFindNode:
				hg.HandleResFindNode(res.(*hashinfocommon.ResFindNode))
			default:
				panic("no way")
			}
		}
	} else {
		log.Warningf("not match res received tid: %v,drop it", helpful.Get10Hex(tid))
	}
}

func (hg *HashInfoGetter) HandleResFindNode(resFindNode *hashinfocommon.ResFindNode) {
	if err := resFindNode.CheckValid(); err != nil {
		log.Warningf("not valid ResFindNode err: %v", err)
	} else {
		nodes := resFindNode.GetNodes()
		//log.Infof("get %v nodes", len(nodes))
		for _, node := range nodes {
			if _, ok := hg.uniqueNodePool[node.Id]; ok {
				//log.Warningf("node repeat id: %v", helpful.GetHex(node.Id))
				continue
			}
			hg.uniqueNodePool[node.Id] = struct{}{}

			hg.FlowControl.WaitFlow()

			if err := hg.SendFindNode(node.Addr, hg.selfId, node.Id[:15]+hg.selfId[15:]); err != nil {
				log.Warningf("send find_node err: %v", err)
			}
		}
	}
}

func (hg *HashInfoGetter) DispatchError(buf []byte, tid string) {
	if _, ok := hg.resPrototypeDict[tid]; ok {
		delete(hg.resPrototypeDict, tid)

		hg.Ins.SafeDo(func() {
			hg.Ins.ReceivedErrors += 1
		})

	} else {
		log.Warningf("received a tid not match res, tid: %v", tid)
	}
}
