package main

import (
	"sync"
	"time"
	"log"
)

type BlockQueue struct {
	queue []interface{}
	mtx   sync.Mutex
	cond  *sync.Cond
}

func newBlockQueue() *BlockQueue {
	blockQueue := &BlockQueue{}
	blockQueue.cond = sync.NewCond(&blockQueue.mtx)
	return blockQueue
}

func (bq *BlockQueue) put(ele interface{}) {
	bq.mtx.Lock()
	defer bq.mtx.Unlock()
	bq.queue = append(bq.queue, ele)
	bq.cond.Signal()
}

func (bq *BlockQueue) take() interface{} {
	bq.mtx.Lock()
	for !(len(bq.queue) > 0) {
		bq.cond.Wait()
	}

	defer bq.mtx.Unlock()
	val := bq.queue[0]
	bq.queue = bq.queue[1:]
	return val
}

func main() {
	bq := newBlockQueue()

	start := time.Now()

	go func() {
		for i := 0; i < 1000000; i++ {
			bq.put(i)
		}
	}()

	for i := 0; i < 1000000; i++ {
		val := bq.take()
		_ = val
	}

	elapsed := time.Since(start)
	log.Printf("took %s", elapsed)
}
