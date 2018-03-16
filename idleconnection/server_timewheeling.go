package main

import (
	"net"
	"container/ring"
	"time"
	"os"
	"fmt"
)

type TimeWheel struct {
	// ring[bucket,bucket,bucket,...]
	// bucket -> set[task,task,task,...] -> map[task]struct{}{}
	slots *ring.Ring

	// the bucket of the ring which specify task last added to
	// map[task]bucket
	lastBucket map[interface{}]interface{}

	durationPerTick time.Duration

	addChan chan interface{}
	delChan chan interface{}

	onTick func(interface{})
}

func New(ticksPerWheel int, durationPerTick time.Duration, f func(interface{})) *TimeWheel {
	if ticksPerWheel < 1 {
		return nil
	}

	tw := &TimeWheel{
		slots:           ring.New(ticksPerWheel),
		durationPerTick: durationPerTick,
		onTick:          f}

	slotsLen := tw.slots.Len()

	// init slot's each bucket with set[task,task,task...]
	for i := 0; i < slotsLen; i++ {
		tw.slots.Value = map[interface{}]struct{}{}
	}

	return tw
}

// two feature
// 1. add new task to TimeWheel
// 2. increase task life in TimeWheel
func (tw *TimeWheel) add(task interface{}) {
	tw.addChan <- task
}

// delete task life bind in TimeWheel
func (tw *TimeWheel) del(task interface{}) {
	tw.delChan <- task
}

// may be run in goroutine
// get all event [add,del,...]
func (tw *TimeWheel) ticksTillDie() {
	ticker := time.NewTicker(tw.durationPerTick)

	for {
		select {
		case task := <-tw.addChan:
			if lastBucket, ok := tw.lastBucket[task]; ok {
				// current bucket's life is longest
				if tw.slots.Value == lastBucket {
					continue
				}

				// delete prev task in time wheeling
				delete(lastBucket.(map[interface{}]struct{}), task)
				delete(tw.lastBucket, task)
			}

			// save task in current bucket
			tw.slots.Value.(map[interface{}]struct{})[task] = struct{}{}
			tw.lastBucket[task] = tw.slots.Value

		case task := <-tw.delChan:
			if lastBucket, ok := tw.lastBucket[task]; ok {
				// delete prev task in time wheeling
				delete(lastBucket.(map[interface{}]struct{}), task)
				delete(tw.lastBucket, task)
			}

		case <-ticker.C:
			// stop tasks' life

			for task, _ := range tw.slots.Value.(map[interface{}]struct{}) {
				delete(tw.slots.Value.(map[interface{}]struct{}), task)
				delete(tw.lastBucket, task)
				tw.onTick(task)
			}
		}
	}
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func serverConn(conn net.Conn) {
	defer conn.Close()

}

func main() {

	//arg := os.Args
	//if len(arg) < 2 {
	//	fmt.Printf("Usage:\n %v listenaddr\nExample:\n %v :10001\n", arg[0], arg[0])
	//	return
	//}
	//
	//listener, err := net.Listen("tcp", arg[1])
	//panicOnError(err)
	//
	//defer listener.Close()
	//
	//for {
	//	localConn, err := listener.Accept()
	//	if err != nil {
	//		continue
	//	}
	//
	//	go serverConn(localConn)
	//}
}
