package common

import (
	"bufio"
	"errors"
	"bytes"
)

type Msg struct {
	Cmd     string
	Topic   string
	Content []byte
}

func (msg *Msg) Serialize() []byte {
	var buf bytes.Buffer
	buf.WriteString(msg.Cmd + " ")
	buf.WriteString(msg.Topic + "\r\n")
	if msg.Content != nil {
		buf.Write(msg.Content)
	}
	return buf.Bytes()
}

// 最大读取字节数设置在bufio.Reader内
func ReadMsg(bufReader *bufio.Reader) (*Msg, error) {

	line, isPrefix, err := bufReader.ReadLine()

	if err != nil || isPrefix {
		return nil, err
	}

	msg := &Msg{}

	if len(line) > 3 && string(line[:3]) == "sub" {
		msg.Cmd = "sub"
		if len(line) > 4 {
			msg.Topic = string(line[4:])
		}
		return msg, nil
	} else if len(line) > 5 && string(line[:5]) == "unsub" {
		msg.Cmd = "ubsub"

		if len(line) > 6 {
			msg.Topic = string(line[6:])
		}
		return msg, nil
	} else if len(line) > 3 && string(line[:3]) == "pub" {
		msg.Cmd = "pub"

		if len(line) > 4 {
			msg.Topic = string(line[4:])
		}

		line, isPrefix, err = bufReader.ReadLine()

		if err != nil || isPrefix {
			return nil, err
		}

		msg.Content = line
		return msg, nil
	}

	return nil, errors.New("invalid msg")
}
