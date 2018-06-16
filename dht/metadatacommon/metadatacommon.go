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
)

// bittorrent message ID
const (
	Extended = 20
)

// extended message ID
const (
	HandShake = 0
)

// ut_metadata?
// msg_type?
const (
	Request = 0
	Data    = 1
	Reject  = 2
)

const (
	MaxBufferLen = 65536
)

type M struct {
	Ut_metadata int `bencode:"ut_metadata"`
}

type ExtHandshake struct {
	M             M   `bencode:"m"`
	Metadata_size int `bencode:"metadata_size"`
}

type RequestPack struct {
	Msg_Type int `bencode:"msg_type"`
	Piece    int `bencode:"piece"`
}

// http://www.bittorrent.org/beps/bep_0010.html
type Packet struct {
	Len int32
	// bittorrent message ID, = 20
	BitMsgId byte
	// extended message ID. 0 = handshake, >0 = extended message as specified by the handshake.
	ExtMsgId byte
	Payload  []byte
}

func (packet *Packet) IsHandShake() bool {
	return packet.ExtMsgId == HandShake
}

func (packet *Packet) IsExtended() bool {
	return packet.BitMsgId == Extended
}

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
			} else if Len > MaxBufferLen {
				return nil, errors.New(fmt.Sprintf("peer msg too long: %v", Len))
			}

			rbuf := make([]byte, Len+4 /*include first4bytes*/)

			rn, err := io.ReadFull(bufReader, rbuf)
			if err != nil || rn != len(rbuf) {
				return nil, err
			}

			packet := &Packet{}
			packet.Len, _ = helpful.ReadInt32(rbuf[:4])
			packet.BitMsgId = rbuf[4]
			packet.ExtMsgId = rbuf[5]
			packet.Payload = rbuf[6:]

			return packet, nil
		}
	}
}

func WriteAPacket(conn net.Conn, bitMsgId, extMsgId byte, payload []byte) error {
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
