package common

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"hash/adler32"
	"github.com/golang/protobuf/proto"
	"fmt"
	"reflect"
	"net"
)

const (
	MaxLen = 16384
)

type Packet struct {
	Len          int32
	NameLen      int32
	TypeName     []byte
	ProtobufData []byte
	CheckSum     int32
}

func readInt32(data []byte) (int32, error) {
	var ret int32
	buf := bytes.NewBuffer(data)
	err := binary.Read(buf, binary.BigEndian, &ret)
	if err != nil {
		return 0, err
	} else {
		return ret, nil
	}
}

func createMessage(typeName string) (proto.Message, error) {
	mt := proto.MessageType(typeName)
	if mt == nil {
		fmt.Errorf("unknown message type %q", typeName)
	}

	return reflect.New(mt.Elem()).Interface().(proto.Message), nil
}

// return proto.Message as interface
func ReadAMessage(bufReader *bufio.Reader) (proto.Message, error) {
	first4bytes, err := bufReader.Peek(4)

	if err != nil {
		return nil, err
	}

	Len, err := readInt32(first4bytes);
	if err != nil || Len > MaxLen {
		return nil, errors.New("first 4 bytes error")
	}

	// 示例代理不优化性能 TODO
	rbuf := make([]byte, Len+4 /*include first4bytes*/)

	rn, err := io.ReadFull(bufReader, rbuf)
	if err != nil || rn != len(rbuf) {
		return nil, err
	}

	cheksum := int32(adler32.Checksum(rbuf[:len(rbuf)-4]))
	_ = cheksum

	packet := &Packet{}
	packet.Len, _ = readInt32(rbuf[:4])
	packet.NameLen, _ = readInt32(rbuf[4:8])
	packet.TypeName = rbuf[8 : 8+packet.NameLen]
	packet.ProtobufData = rbuf[8+packet.NameLen : len(rbuf)-4]
	packet.CheckSum, _ = readInt32(rbuf[len(rbuf)-4:])

	// don't check
	if cheksum != packet.CheckSum {
		return nil, errors.New("checksum error")
	}

	// TypeName 必须是\0末尾的
	var TypeName string
	if len(packet.TypeName) > 0 && (packet.TypeName[len(packet.TypeName)-1] == 0) {
		TypeName = string(packet.TypeName[:len(packet.TypeName)-1])
	} else {
		return nil, errors.New("TypeName error")
	}

	message, err := createMessage(TypeName)
	if err != nil {
		return nil, err
	}

	err = proto.Unmarshal(packet.ProtobufData, message)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func WriteAMessage(conn net.Conn, message proto.Message) error {
	var typeName = proto.MessageName(message) + string([]byte{0})
	var nameLen = int32(len(typeName))
	var protobufLen = int32(proto.Size(message))
	var checkSum int32

	var len int32 = 4 /*namelen*/ + nameLen + protobufLen + 4 /*checksum*/

	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, &len)
	binary.Write(&buf, binary.BigEndian, &nameLen)
	buf.Write([]byte(typeName))

	marshaled, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	buf.Write(marshaled)
	checkSum = int32(adler32.Checksum(buf.Bytes()))

	binary.Write(&buf, binary.BigEndian, &checkSum)

	wn, err := conn.Write(buf.Bytes())
	if err != nil {
		return err
	}

	_ = wn

	return nil
}
