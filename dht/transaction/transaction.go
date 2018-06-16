package transaction

import (
	"sync/atomic"
	"encoding/binary"
)

type Transaction struct {
	Id uint64
}

// 他的协议看似是字符串,其实[]byte,保存的一个一个字节

// 0 ~ 65535
func (tm *Transaction) FetchAndAdd() string {
	id := atomic.AddUint64(&tm.Id, 1) - 1
	var buf [2]byte
	binary.BigEndian.PutUint64(buf[:], id%65536)
	return string(buf[:])
}
