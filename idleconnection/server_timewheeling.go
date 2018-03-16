package main

import (
	"net"
	"container/ring"
	"time"
	"log"
	"os"
	"fmt"
	"io"
)

//  map[ele]struct{} like c++'s unordered_set
type Bucket struct {
	eles map[interface{}]struct{}
}

func newBucket() *Bucket {
	bucket := Bucket{}
	bucket.eles = make(map[interface{}]struct{})
	return &bucket
}

func (bucket *Bucket) deleteEle(ele interface{}) {
	delete(bucket.eles, ele)
}

func (bucket *Bucket) addEle(ele interface{}) {
	bucket.eles[ele] = struct{}{}
}

type TimeWheel struct {
	// ring[bucket*,bucket*,bucket*,...]
	slots *ring.Ring

	// the bucket of the ring which specify ele last added to
	// map[ele]bucket*
	lastBucket map[interface{}]*Bucket

	durationPerTick time.Duration

	addChan chan interface{}
	delChan chan interface{}

	onTick func(interface{})

	// when ticket print whole wheel
	debugLog bool

	firstSlot *ring.Ring
}

func New(ticksPerWheel int, durationPerTick time.Duration, f func(interface{})) *TimeWheel {
	if ticksPerWheel < 1 {
		return nil
	}

	tw := &TimeWheel{
		slots:           ring.New(ticksPerWheel),
		lastBucket:      make(map[interface{}]*Bucket),
		durationPerTick: durationPerTick,
		addChan:         make(chan interface{}),
		delChan:         make(chan interface{}),
		onTick:          f,
		debugLog:        false,
		firstSlot:       nil}

	// init firstSlot for debug print
	tw.firstSlot = tw.slots

	slotsLen := tw.slots.Len()

	// init slot's each bucket with Bucket
	for i := 0; i < slotsLen; i++ {
		tw.slots.Value = newBucket()
		tw.slots = tw.slots.Next()
	}

	return tw
}

// two feature
// 1. add new ele to TimeWheel
// 2. increase ele's life in TimeWheel
func (tw *TimeWheel) add(ele interface{}) {
	tw.addChan <- ele
}

// delete ele life bind in TimeWheel
func (tw *TimeWheel) del(ele interface{}) {
	tw.delChan <- ele
}

func (tw *TimeWheel) lifeLongestBucket() *Bucket {
	return tw.slots.Prev().Value.(*Bucket)
}

func (tw *TimeWheel) currentBucket() *Bucket {
	return tw.slots.Value.(*Bucket)
}

// may be run in goroutine
// get all event [add,del,...]
func (tw *TimeWheel) ticksTillDie() {
	ticker := time.NewTicker(tw.durationPerTick)

	for {
		select {
		case ele := <-tw.addChan:
			if lastBucket, ok := tw.lastBucket[ele]; ok {
				// pre bucket's life is longest
				if lastBucket == tw.lifeLongestBucket() {
					continue
				}

				// delete prev ele in time wheeling
				lastBucket.deleteEle(ele)
				delete(tw.lastBucket, ele)
			}

			// save ele in longest life's bucket
			tw.lifeLongestBucket().addEle(ele)
			tw.lastBucket[ele] = tw.lifeLongestBucket()

		case ele := <-tw.delChan:
			if lastBucket, ok := tw.lastBucket[ele]; ok {
				// delete prev ele in time wheeling
				lastBucket.deleteEle(ele)
				delete(tw.lastBucket, ele)
			}

		case <-ticker.C:
			if tw.debugLog {
				n := 0
				tw.firstSlot.Do(func(bucketInterface interface{}) {
					bucket := bucketInterface.(*Bucket)
					symbol := ""
					if bucket == tw.currentBucket() {
						symbol = "<-"
					}
					log.Printf("[%v] len = %v %v\n", n, len(bucket.eles), symbol)
					n += 1
				})
				log.Printf("%v", "===========================================")
			}

			// stop eles' life
			for ele, _ := range tw.currentBucket().eles {
				tw.currentBucket().deleteEle(ele)
				delete(tw.lastBucket, ele)
				tw.onTick(ele)
			}

			tw.slots = tw.slots.Next()
		}
	}
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func serverConn(conn net.Conn, tw *TimeWheel) {
	defer conn.Close()
	tw.add(conn)

	go io.Copy(conn, conn)
}

func main() {

	arg := os.Args
	if len(arg) < 2 {
		fmt.Printf("Usage:\n %v listenaddr\nExample:\n %v :10001\n", arg[0], arg[0])
		return
	}

	// timer wheel
	tw := New(10, time.Second*1, func(ele interface{}) {
		conn := ele.(net.Conn)
		conn.Close()
	})
	tw.debugLog = true

	go func(tw *TimeWheel) {
		tw.ticksTillDie()
	}(tw)

	// server
	listener, err := net.Listen("tcp", arg[1])
	panicOnError(err)

	defer listener.Close()

	for {
		localConn, err := listener.Accept()
		if err != nil {
			continue
		}

		go serverConn(localConn, tw)
	}
}
