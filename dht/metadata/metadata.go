package metadata

import (
	"github.com/yqsy/recipes/dht/inspector"
	"net"
	"github.com/Sirupsen/logrus"
	"io"
	"bytes"
	"github.com/yqsy/recipes/dht/helpful"
	"errors"
	"github.com/yqsy/recipes/dht/metadatacommon"
	"github.com/zeebo/bencode"
)

const (
	BLOCK = 16384
)

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

// 这段没有找到文档
func (mg *MetaGetter) HandleShake(conn net.Conn, metaSource *MetaSource) error {
	var sendBuf bytes.Buffer
	sendBuf.Write(handshakePrefix)             // 28
	sendBuf.Write([]byte(metaSource.Hashinfo)) // 20
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

func (mg *MetaGetter) ExtHandleShake(conn net.Conn, metaSource *MetaSource) (*metadatacommon.ExtHandshake, error) {
	extHandShake := &metadatacommon.ExtHandshake{M: metadatacommon.M{Ut_metadata: metadatacommon.Data}}
	if reqBytes, err := bencode.EncodeBytes(extHandShake); err != nil {
		return nil, err
	} else {
		if err = metadatacommon.WriteAPacket(conn, metadatacommon.Extended, metadatacommon.HandShake, reqBytes); err != nil {
			return nil, err
		}

		resPacket, err := metadatacommon.ReadAPacket(conn)
		if err != nil {
			return nil, err
		}

		if !resPacket.IsExtended() || !resPacket.IsHandShake() {
			return nil, errors.New("extHandshake error")
		}

		extHandShakeRes := &metadatacommon.ExtHandshake{}
		if err := bencode.DecodeBytes(resPacket.Payload, extHandShakeRes); err != nil {
			return nil, err
		}

		return extHandShakeRes, nil
	}
}

func (mg *MetaGetter) GetPieces(conn net.Conn, extHandshakeRes *metadatacommon.ExtHandshake) error {
	piecesNum := extHandshakeRes.Metadata_size / BLOCK
	if extHandshakeRes.Metadata_size%BLOCK != 0 {
		piecesNum++
	}

	for i := 0; i < piecesNum; i++ {
		request := metadatacommon.RequestPack{Msg_Type: metadatacommon.Request, Piece: i}

		if reqBytes, err := bencode.EncodeBytes(request); err != nil {
			return err
		} else {
			if err = metadatacommon.WriteAPacket(conn, metadatacommon.Extended, byte(extHandshakeRes.M.Ut_metadata), reqBytes); err != nil {
				return err
			} else {

				resPacket, err := metadatacommon.ReadAPacket(conn)
				if err != nil {
					return err
				}

				_ = resPacket

				// TODO fuck 什么鬼协议
			}
		}
	}

	return nil
}

func (mg *MetaGetter) Serve(conn net.Conn, metaSource *MetaSource) {
	defer conn.Close()

	for {
		if err := mg.HandleShake(conn, metaSource); err != nil {
			logrus.Warnf("handshake err: %v", err)
			break
		}

		extHandshakeRes, err := mg.ExtHandleShake(conn, metaSource)
		if err != nil {
			logrus.Warnf("extHandleShake err: %v", err)
			break
		}

		err = mg.GetPieces(conn, extHandshakeRes)
		if err != nil {
			logrus.Warnf("GetPieces err: %v", err)
			break
		}
	}
}

func (mg *MetaGetter) Run(metaSourceChan chan *MetaSource) error {

	for {
		metaSource := <-metaSourceChan
		if conn, err := net.Dial("tcp", metaSource.Addr); err != nil {
			logrus.Warnf("connect %v err: %v", metaSource.Addr, err)
		} else {
			go mg.Serve(conn, metaSource)
		}
	}

	return nil
}
