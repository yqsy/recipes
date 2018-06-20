package transaction

import (
	"sync/atomic"
)

type Transaction struct {
	Id uint64
}

// BigEndian
func int2bytes(val uint64) []byte {
	data := make([]byte, 8)
	j := -1

	for i := 0; i < 8; i++ {
		shift := uint64((7 - i) * 8)
		data[i] = byte((val & (0xff << shift)) >> shift)

		if j == -1 && data[i] != 0 {
			j = i
		}
	}

	if j == -1 {
		// if val is zero
		return data[:1]
	} else {
		// remove all zero at bigendian side
		return data[j:]
	}
}

// 他的协议看似是字符串,其实[]byte,保存的一个一个字节
func (tm *Transaction) FetchAndAdd() string {
	id := atomic.AddUint64(&tm.Id, 1)
	return string(int2bytes(id))
}
