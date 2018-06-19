package metadatacommon

import (
	"net"
	"bytes"
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"fmt"
	"github.com/yqsy/recipes/dht/helpful"
	"reflect"
)

// bittorrent message ID
const (
	Extended = 20
)

// extended message ID
const (
	HandShake = 0
)

// msg_type?
const (
	Request = 0
	Data    = 1
	Reject  = 2
)

const (
	MaxBufferLen = 65536
)

// http://www.bittorrent.org/beps/bep_0010.html
type ExtPacket struct {
	Len int32
	// bittorrent message ID, = 20
	BitMsgId byte
	// extended message ID. 0 = handshake, >0 = extended message as specified by the handshake.
	ExtMsgId byte
	Payload  []byte
}

type Packet struct {
	Len      int32
	BitMsgId byte
	Payload  []byte
}

// 冗余一点,没关系了,这个协议太恶心了
func ReadAPacket(conn net.Conn) (*Packet, error) {
	bufReader := bufio.NewReader(conn)
	if first4bytes, err := bufReader.Peek(4); err != nil {
		return nil, err
	} else {
		if Len, err := helpful.ReadInt32(first4bytes); err != nil {
			return nil, err
		} else {

			// http://jonas.nitro.dk/bittorrent/bittorrent-rfc.html 6.2
			// If a message has no payload, its size is 1.
			// Messages of size 0 MAY be sent periodically as keep-alive messages.
			if Len == 1 {
				return nil, errors.New("msg size is 1, no payload")
			} else if Len == 0 {
				return nil, errors.New("msg size is 0, keep-live? ")
			} else if Len > MaxBufferLen || Len < 0 {
				return nil, errors.New(fmt.Sprintf("peer msg len error: %v", Len))
			}

			rbuf := make([]byte, Len+4 /*include first4bytes*/)

			rn, err := io.ReadFull(bufReader, rbuf)
			if err != nil || rn != len(rbuf) {
				return nil, err
			}

			packet := &Packet{}
			packet.Len, _ = helpful.ReadInt32(rbuf[:4])
			packet.BitMsgId = rbuf[4]
			packet.Payload = rbuf[5:]

			return packet, nil
		}
	}
}

func (ep *ExtPacket) IsHandShake() bool {
	return ep.ExtMsgId == HandShake
}

func (ep *ExtPacket) IsExtended() bool {
	return ep.BitMsgId == Extended
}

func ReadAExtPacket(conn net.Conn) (*ExtPacket, error) {
	bufReader := bufio.NewReader(conn)
	if first4bytes, err := bufReader.Peek(4); err != nil {
		return nil, err
	} else {
		if Len, err := helpful.ReadInt32(first4bytes); err != nil {
			return nil, err
		} else {

			// http://jonas.nitro.dk/bittorrent/bittorrent-rfc.html 6.2
			// If a message has no payload, its size is 1.
			// Messages of size 0 MAY be sent periodically as keep-alive messages.
			if Len == 1 {
				return nil, errors.New("msg size is 1, no payload")
			} else if Len == 0 {
				return nil, errors.New("msg size is 0, keep-live? ")
			} else if Len > MaxBufferLen || Len < 0 {
				return nil, errors.New(fmt.Sprintf("peer msg len error: %v", Len))
			}

			rbuf := make([]byte, Len+4 /*include first4bytes*/)

			rn, err := io.ReadFull(bufReader, rbuf)
			if err != nil || rn != len(rbuf) {
				return nil, err
			}

			packet := &ExtPacket{}
			packet.Len, _ = helpful.ReadInt32(rbuf[:4])
			packet.BitMsgId = rbuf[4]
			packet.ExtMsgId = rbuf[5]
			packet.Payload = rbuf[6:]

			return packet, nil
		}
	}
}

func WriteAExtPacket(conn net.Conn, bitMsgId, extMsgId byte, payload []byte) error {
	var Len int32 = 2 + int32(len(payload))

	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, &Len)
	buf.Write([]byte{bitMsgId, extMsgId})
	buf.Write(payload)

	if wn, err := conn.Write(buf.Bytes()); wn != buf.Len() || err != nil {
		return err
	} else {
		return nil
	}
}

func CheckExtHandShakeRes(b interface{}) error {
	if obj, ok := b.(map[string]interface{}); !ok {
		return errors.New("not an obj")
	} else {
		if m, ok := obj["m"]; !ok ||
			reflect.TypeOf(m).Kind() != reflect.Map {
			return errors.New("m error")
		} else {
			if ut_metadata, ok := m.(map[string]interface{})["ut_metadata"]; !ok ||
				reflect.TypeOf(ut_metadata).Kind() != reflect.Int {
				return errors.New("m.ut_metadata error")
			}
		}

		if metadata_size, ok := obj["metadata_size"]; !ok ||
			reflect.TypeOf(metadata_size).Kind() != reflect.Int {
			return errors.New("metadata_size error")
		}
	}

	return nil
}
