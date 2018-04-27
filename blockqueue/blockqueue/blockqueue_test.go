package blockqueue

import (
	"time"
	"testing"
)

const (
	SwitchTimes = 10000000
)

func TestSimple(t *testing.T) {
	bq := NewBlockQueue()

	start := time.Now()

	go func() {
		for i := 0; i < SwitchTimes; i++ {
			bq.Put(i)
		}
	}()

	for i := 0; i < SwitchTimes; i++ {
		val := bq.Take()
		_ = val
	}

	elapsed := time.Since(start)
	t.Logf("SwitchTimes:%v took:%v speed:%.2f/s", SwitchTimes, elapsed, SwitchTimes/(elapsed.Seconds()))
}
