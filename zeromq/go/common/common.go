package common

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"net"
)

type Global struct {
	MsgSize int

	RoundTripCount int
}

type Pack struct {
	Len int32

	Body []byte
}

type Context struct {
	Conn net.Conn

	BufReader *bufio.Reader

	// 读的是定长的
	ReadBuf []byte
}

// body 是来自context的ReadBuf的引用
func (ctx *Context) ReadPackage() (*Pack, error) {

	pack := &Pack{}

	err := binary.Read(ctx.BufReader, binary.BigEndian, &pack.Len)

	if err != nil {
		return nil, err
	}

	if pack.Len > 65536 || pack.Len < 0 {
		return nil, errors.New("invalid length")
	}

	if int(pack.Len) != len(ctx.ReadBuf) {
		return nil, errors.New("length not equal")
	}

	rn, err := io.ReadFull(ctx.BufReader, ctx.ReadBuf)

	//  pack直接使用ctx的buf的引用, 一次申请即可
	pack.Body = ctx.ReadBuf

	if err != nil || rn != int(pack.Len) {
		return nil, errors.New("invalid body")
	}

	return pack, nil
}
