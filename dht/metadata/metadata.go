package metadata

import (
	"github.com/yqsy/recipes/dht/inspector"
	"net"
	"io"
	"bytes"
	"github.com/yqsy/recipes/dht/helpful"
	"errors"
	"github.com/yqsy/recipes/dht/metadatacommon"
	"github.com/op/go-logging"
	"time"
	"fmt"
	"github.com/yqsy/recipes/dht/bencode"
)

var log = logging.MustGetLogger("dht")

const (
	BLOCK = 16384
)

// Protocol Name Length: 19
// Protocol Name: BitTorrent protocol
// Reserved Extension Bytes: 0000000000100001
var handshakePrefix = []byte{
	19, 66, 105, 116, 84, 111, 114, 114, 101, 110, 116, 32, 112, 114,
	111, 116, 111, 99, 111, 108, 0, 0, 0, 0, 0, 16, 0, 1,
}

type MetaSource struct {
	Hashinfo string
	Addr     string
}

type MetaGetter struct {
	Ins *inspector.Inspector
}

// 这段没有找到文档,只能靠抓包猜测
func (mg *MetaGetter) HandleShake(conn net.Conn, metaSource *MetaSource) error {
	var sendBuf bytes.Buffer
	sendBuf.Write(handshakePrefix) // 28
	// SHA1 Hash of info dictionary: 3dec76b035c3beef644bcb8eaca88406527fa422
	sendBuf.Write([]byte(metaSource.Hashinfo)) // 20
	// Peer ID: bd15803eb26ad29bee2ba9c2b258bcf3d2abf6b1
	sendBuf.Write([]byte(helpful.RandomString(20)))

	if wn, err := conn.Write(sendBuf.Bytes()); err != nil || wn != sendBuf.Len() {
		return err
	}

	var recvBuf [68]byte
	if rn, err := io.ReadFull(conn, recvBuf[:]); rn != len(recvBuf) || err != nil {
		return err
	}

	if !bytes.Equal(handshakePrefix[:20], recvBuf[:20]) {
		return errors.New("handshake response err")
	}

	return nil
}

// req:
// {
//	  m: {
//	      "ut_metadata": 1
//  	}
// }

// res:
//{
//	m: {
//	    "ut_metadata": int   // 拿来下次请求时用
//	}
//	"metadata_size": int  // 分割piece块,请求piece
//}
// http://www.bittorrent.org/beps/bep_0010.html
func (mg *MetaGetter) ExtHandleShake(conn net.Conn, metaSource *MetaSource) (map[string]interface{}, error) {
	extHandShake := map[string]interface{}{
		"m": map[string]interface{}{
			"ut_metadata": 1,
		},
	}

	reqBytes := []byte(bencode.Encode(extHandShake))
	if err := metadatacommon.WriteAExtPacket(conn, metadatacommon.Extended, metadatacommon.HandShake, reqBytes); err != nil {
		return nil, err
	}

	resPacket, err := metadatacommon.ReadAExtPacket(conn)
	if err != nil {
		return nil, err
	}

	if !resPacket.IsExtended() || !resPacket.IsHandShake() {
		return nil, errors.New("extHandshake error")
	}

	if v, err := bencode.Decode(string(resPacket.Payload)); err != nil {
		return nil, errors.New("extHandshake decode error")
	} else {
		if err = metadatacommon.CheckExtHandShakeRes(v); err != nil {
			return nil, err
		} else {
			return v.(map[string]interface{}), nil
		}
	}
}

// http://www.bittorrent.org/beps/bep_0003.html
func (mg *MetaGetter) GetBitField(conn net.Conn) ([]byte /*payload bitmap*/ , error) {
	resPacket, err := metadatacommon.ReadAPacket(conn)

	if err != nil {
		return []byte{}, err
	}

	if resPacket.BitMsgId != 5 {
		return []byte{}, errors.New("message type is not 5")
	}

	return resPacket.Payload, err
}

// http://www.bittorrent.org/beps/bep_0009.html
func (mg *MetaGetter) GetPieces(conn net.Conn, extHandshakeRes map[string]interface{}) error {
	utMetadata := extHandshakeRes["m"].(map[string]interface{})["ut_metadata"].(int)
	metadataSize := extHandshakeRes["metadata_size"].(int)

	piecesNum := metadataSize / BLOCK
	if metadataSize%BLOCK != 0 {
		piecesNum++
	}

	for i := 0; i < piecesNum; i++ {
		requset := map[string]interface{}{
			"msg_type": metadatacommon.Request,
			"piece":    i,
		}

		reqBytes := []byte(bencode.Encode(requset))

		if err := metadatacommon.WriteAExtPacket(conn, metadatacommon.Extended, byte(utMetadata), reqBytes); err != nil {
			return err
		} else {

			for {
				resPacket, err := metadatacommon.ReadAExtPacket(conn)

				if err != nil {
					return nil
				}

				_ = resPacket

			}
		}
	}
	return nil
}

// 这个只是我的猜测
func (mg *MetaGetter) DmdDiscard(conn net.Conn) error {

	for {
		resPacket, err := metadatacommon.ReadAPacket(conn)

		if err != nil {
			return err
		}

		if resPacket.BitMsgId == 4 {
			// nice
		} else if resPacket.BitMsgId == 9 {
			return nil
		} else {
			return errors.New(fmt.Sprintf("error msg id = %v", resPacket.BitMsgId))
		}
	}
}

func (mg *MetaGetter) Serve(conn net.Conn, metaSource *MetaSource) {
	defer conn.Close()

	if err := mg.HandleShake(conn, metaSource); err != nil {
		log.Warningf("handshake err: %v remoteaddr: %v", err, conn.RemoteAddr())
		return
	}

	extHandshakeRes, err := mg.ExtHandleShake(conn, metaSource)
	if err != nil {
		log.Warningf("extHandleShake err: %v remoteaddr: %v", err, conn.RemoteAddr())
		return
	}

	//  discard Message Type: Bitfield (5)
	_, err = mg.GetBitField(conn)
	if err != nil {
		log.Warningf("get bitfield err: %v remoteaddr: %v", err, conn.RemoteAddr())
		return
	}

	// discard Message Type: Have (4) / Message Type: Port (9)
	if err = mg.DmdDiscard(conn); err != nil {
		log.Warningf("dmddiscard err: %v remoteaddr: %v", err, conn.RemoteAddr())
		return
	}

	err = mg.GetPieces(conn, extHandshakeRes)
	if err != nil {
		log.Warningf("GetPieces err: %v remoteaddr: %v", err, conn.RemoteAddr())
		return
	}

	log.Infof("ok????")
}

func (mg *MetaGetter) Run(metaSourceChan chan *MetaSource) error {

	for {
		metaSource := <-metaSourceChan

		if conn, err := net.DialTimeout("tcp", metaSource.Addr, time.Second*15); err != nil {
			log.Warningf("connect %v err", metaSource.Addr)
		} else {
			go mg.Serve(conn, metaSource)
		}
	}

	return nil
}
