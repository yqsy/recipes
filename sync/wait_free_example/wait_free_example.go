package main

import (
	"sync/atomic"
	"runtime"
	"fmt"
)

type TicketStore struct {
	ticket *uint64
	done   *uint64
	slots  [1024]string
}

func New() TicketStore {
	return TicketStore{
		ticket: new(uint64),
		done:   new(uint64)}
}

func (ts *TicketStore) Put(s string) {
	t := atomic.AddUint64(ts.ticket, 1) - 1
	ts.slots[t] = s
	for !atomic.CompareAndSwapUint64(ts.done, t, t+1) {
		runtime.Gosched()
	}
}

func (ts *TicketStore) GetDone() []string {
	return ts.slots[:atomic.LoadUint64(ts.done) /*+1???*/ ]
}

func main() {
	tickStore := New()

	tickStore.Put("1")
	tickStore.Put("2")
	tickStore.Put("3")

	fmt.Println(tickStore.GetDone())
}
