package helpful

import (
	"crypto/rand"
	"fmt"
	"bytes"
	"encoding/binary"
	"strconv"
)

func RandomString(len int) string {
	buf := make([]byte, len)
	rand.Read(buf)
	return string(buf)
}

// 他的协议的表达形式看似是json字符串.
// 其实不是,是数字!!方便人辨识要把内存中的数字转换成人眼可读的数字

func GetHex(str string) (rtn string) {
	buf := []byte(str)
	for i := 0; i < len(buf); i++ {
		rtn += fmt.Sprintf("%02x", buf[i])
	}
	return rtn
}

// str 是大端法的字节流 (至少我生成的tid是)
func Get10Hex(str string) (rtn string) {
	buf := bytes.NewBuffer([]byte(str))

	var ret int16
	err := binary.Read(buf, binary.BigEndian, &ret)
	if err != nil {
		return "-1"
	} else {
		return strconv.Itoa(int(ret))
	}
}

func ReadInt32(data []byte) (int32, error) {
	var ret int32
	buf := bytes.NewBuffer(data)
	err := binary.Read(buf, binary.BigEndian, &ret)
	if err != nil {
		return 0, err
	} else {
		return ret, nil
	}
}
