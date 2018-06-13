package transaction

import (
	"sync/atomic"
)

type Transaction struct {
	Id uint64
}

// 0 ~ 65535
func (tm *Transaction) FetchAndAdd() string {
	id := atomic.AddUint64(&tm.Id, 1) - 1
	return string(id % 65536)
}
