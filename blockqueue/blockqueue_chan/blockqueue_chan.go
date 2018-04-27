package main

import (
	"time"
	"fmt"
)

const (
	SwitchTimes = 10000000
)

func main() {

	queue := make(chan int)

	start := time.Now()

	go func() {
		for i := 0; i < SwitchTimes; i++ {
			queue <- i
		}
	}()

	for i := 0; i < SwitchTimes; i++ {
		val := <-queue
		_ = val
	}

	elapsed := time.Since(start)
	fmt.Printf("SwitchTimes:%v took:%v speed:%.2f/s\n", SwitchTimes, elapsed, SwitchTimes/(elapsed.Seconds()))
}
