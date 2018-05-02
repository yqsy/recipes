package common

import (
	"net"
	"encoding/binary"
	"errors"
	"io"
)

const (
	MaxBufSize = 16384
)

type Message struct {
	Len  int32
	Body []byte
}

func WriteToConnRaw(conn net.Conn, buf []byte) error {
	msg := &Message{}
	msg.Len = int32(len(buf))
	msg.Body = buf

	return WriteToConn(conn, msg)
}

func WriteToConn(conn net.Conn, msg *Message) error {
	err := binary.Write(conn, binary.BigEndian, &msg.Len)

	if err != nil {
		return err
	}

	wn, err := conn.Write(msg.Body)
	if err != nil || wn != len(msg.Body) {
		return err
	}

	return nil
}

func ReadFromConn(conn net.Conn) (*Message, error) {
	msg := &Message{}
	err := binary.Read(conn, binary.BigEndian, &msg.Len)
	if err != nil {
		return nil, err
	}

	if msg.Len > MaxBufSize {
		return nil, errors.New("message too big")
	}

	buf := make([]byte, msg.Len)

	rn, err := io.ReadFull(conn, buf)
	if err != nil || rn != len(buf) {
		return nil, err
	}

	msg.Body = buf

	return msg, nil
}
