package main

import (
	"sync/atomic"
	"runtime"
	"time"
	"fmt"
)

type SpinLock struct {
	state *int32
}

func New() SpinLock {
	return SpinLock{state: new(int32)}
}

const free = int32(0)

func (l *SpinLock) Lock() {
	for !atomic.CompareAndSwapInt32(l.state, free, 42) {
		runtime.Gosched()
	}
}

func (l *SpinLock) Unlock() {
	atomic.StoreInt32(l.state, free)
}

func main() {
	spinLock := New()

	locked := make(chan bool, 1)

	go func(spinLock SpinLock, locked chan bool) {
		spinLock.Lock()
		defer spinLock.Unlock()
		locked <- true
		time.Sleep(time.Second * 5)
	}(spinLock, locked)

	<-locked
	spinLock.Lock()
	defer spinLock.Unlock()
	fmt.Println("done")
}
