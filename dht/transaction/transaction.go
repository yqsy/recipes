package transaction

import (
	"sync/atomic"
)

type Transaction struct {
	Id uint64
}

// int2bytes returns the byte array it represents.
func int2bytes(val uint64) []byte {
	data, j := make([]byte, 8), -1
	for i := 0; i < 8; i++ {
		shift := uint64((7 - i) * 8)
		data[i] = byte((val & (0xff << shift)) >> shift)

		if j == -1 && data[i] != 0 {
			j = i
		}
	}

	if j != -1 {
		return data[j:]
	}
	return data[:1]
}

// 他的协议看似是字符串,其实[]byte,保存的一个一个字节
func (tm *Transaction) FetchAndAdd() string {
	id := atomic.AddUint64(&tm.Id, 1)
	return string(int2bytes(id))
}
