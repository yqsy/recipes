package main

import (
	"time"
	"log"
)

func main() {

	queue := make(chan int)

	start := time.Now()

	go func() {
		for i := 0; i < 1000000; i++ {
			queue <- i
		}
	}()

	for i := 0; i < 1000000; i++ {
		val := <-queue
		_ = val
	}

	elapsed := time.Since(start)
	log.Printf("took %s", elapsed)
}
