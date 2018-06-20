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
	"github.com/yqsy/recipes/dht/bencode"
	"time"
	"crypto/sha1"
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
	Infohash string
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
	sendBuf.Write([]byte(metaSource.Infohash)) // 20
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

// write all pieces request
func (mg *MetaGetter) WritePiecesReq(conn net.Conn, piecesNum, utMetadata int) error {
	for i := 0; i < piecesNum; i++ {
		req := map[string]interface{}{
			"msg_type": metadatacommon.Request,
			"piece":    i,
		}

		reqBytes := []byte(bencode.Encode(req))

		if err := metadatacommon.WriteAExtPacket(conn, metadatacommon.Extended, byte(utMetadata), reqBytes); err != nil {
			return err
		}
	}
	return nil
}

func (mg *MetaGetter) ReadPiecesRes(conn net.Conn, piecesNum int, metadataSize int) ([][]byte, error) {
	// get all pieces response
	pieces := make([][]byte, piecesNum)
	for i := 0; i < len(pieces); {
		if resPacket, err := metadatacommon.ReadAExtPacket(conn); err != nil {
			return nil, err
		} else {
			if resPacket.BitMsgId != metadatacommon.Extended ||
				resPacket.ExtMsgId != metadatacommon.Data {
			} else {
				if res, remain, err := bencode.DecodeAndLeak(string(resPacket.Payload)); err != nil {
					return nil, err
				} else {
					if err = metadatacommon.CheckPieceRes(res); err != nil {
						return nil, err
					}
					if res.(map[string]interface{})["msg_type"].(int) != metadatacommon.Data {
						return nil, errors.New("msg_type != Data")
					}
					pieceIdx := res.(map[string]interface{})["piece"].(int)
					if pieceIdx > len(pieces) {
						return nil, errors.New("piece out of range")
					}

					// the last piece
					if pieceIdx == piecesNum-1 && len(remain) != metadataSize/BLOCK {
						return nil, errors.New("last piece error")
					} else if len(remain) != BLOCK {
						return nil, errors.New("piece error")
					}
					pieces[pieceIdx] = []byte(remain)
					i++
				}
			}
		}
	}

	for i := 0; i < len(pieces); i++ {
		if len(pieces[i]) == 0 {
			return nil, errors.New("pieces not complete")
		}
	}

	return pieces, nil
}

// http://www.bittorrent.org/beps/bep_0009.html
func (mg *MetaGetter) GetMetadata(conn net.Conn, extHandshakeRes map[string]interface{}, infoHash string) (interface{}, error) {
	utMetadata := extHandshakeRes["m"].(map[string]interface{})["ut_metadata"].(int)
	metadataSize := extHandshakeRes["metadata_size"].(int)
	piecesNum := metadataSize / BLOCK
	if metadataSize%BLOCK != 0 {
		piecesNum++
	}

	if err := mg.WritePiecesReq(conn, piecesNum, utMetadata); err != nil {
		return nil, err
	}

	if pieces, err := mg.ReadPiecesRes(conn, piecesNum, metadataSize); err != nil {
		return nil, err
	} else {
		metaDataInfo := bytes.Join(pieces, nil)
		sha1sum := sha1.Sum(metaDataInfo)
		if !bytes.Equal([]byte(infoHash), sha1sum[:]) {
			return nil, errors.New("sha1 not equal")
		}

		if v, err := bencode.Decode(string(metaDataInfo)); err != nil {
			return nil, errors.New("decode metaDataInfo error")
		} else {
			return v, nil
		}
	}
}

func (mg *MetaGetter) Serve(conn net.Conn, metaSource *MetaSource) {
	defer conn.Close()

	if err := mg.HandleShake(conn, metaSource); err != nil {
		log.Warningf("handshake err: %v remoteAddr: %v", err, conn.RemoteAddr())
		return
	}

	// get utMetadata and metadataSize
	extHandshakeRes, err := mg.ExtHandleShake(conn, metaSource)
	if err != nil {
		log.Warningf("extHandleShake err: %v remoteAddr: %v", err, conn.RemoteAddr())
		return
	}

	metaData, err := mg.GetMetadata(conn, extHandshakeRes, metaSource.Infohash)
	if err != nil {
		log.Warningf("GetMetadata err: %v remoteAddr: %v", err, conn.RemoteAddr())
		return
	}

	log.Infof(bencode.Prettify(metaData))
}

func (mg *MetaGetter) Run(metaSourceChan chan *MetaSource) error {

	for {
		metaSource := <-metaSourceChan

		log.Infof("infohash: %v remote: %v", helpful.GetHex(metaSource.Infohash), metaSource.Addr)

		if conn, err := net.DialTimeout("tcp", metaSource.Addr, time.Second*15); err != nil {
			log.Warningf("connect %v err", metaSource.Addr)
		} else {

			go mg.Serve(conn, metaSource)
		}
	}

	return nil
}
