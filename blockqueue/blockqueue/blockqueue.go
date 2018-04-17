package blockqueue

import (
	"sync"
)

type BlockQueue struct {
	queue []interface{}
	mtx   sync.Mutex
	cond  *sync.Cond
}

func NewBlockQueue() *BlockQueue {
	blockQueue := &BlockQueue{}
	blockQueue.cond = sync.NewCond(&blockQueue.mtx)
	return blockQueue
}

func (bq *BlockQueue) Put(ele interface{}) {
	bq.mtx.Lock()
	defer bq.mtx.Unlock()
	bq.queue = append(bq.queue, ele)
	bq.cond.Signal()
}

func (bq *BlockQueue) Take() interface{} {
	bq.mtx.Lock()
	for !(len(bq.queue) > 0) {
		bq.cond.Wait()
	}

	defer bq.mtx.Unlock()
	val := bq.queue[0]
	bq.queue = bq.queue[1:]
	return val
}
