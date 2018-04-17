package blockqueue

import (
	"time"
	"testing"
)

func TestSimple(t *testing.T) {
	bq := NewBlockQueue()

	start := time.Now()

	go func() {
		for i := 0; i < 1000000; i++ {
			bq.Put(i)
		}
	}()

	for i := 0; i < 1000000; i++ {
		val := bq.Take()
		_ = val
	}

	elapsed := time.Since(start)
	t.Logf("took %s", elapsed)
}
