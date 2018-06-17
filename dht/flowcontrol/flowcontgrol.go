package flowcontrol

import (
	"sync"
	"time"
)

type FlowControl struct {
	mtx sync.Mutex

	cond *sync.Cond

	// 可使用流量
	flow int
}

func NewFlowControl() *FlowControl {
	fc := &FlowControl{}
	fc.cond = sync.NewCond(&fc.mtx)
	return fc
}

func (fc *FlowControl) Increasing(flowPerSecond int) {
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ticker.C:
			fc.mtx.Lock()
			// 不累积了,每s最多数字
			fc.flow = flowPerSecond
			fc.mtx.Unlock()
			fc.cond.Signal()
		}
	}
}

func (fc *FlowControl) WaitFlow() {
	fc.mtx.Lock()

	if !(fc.flow > 0) {
		fc.cond.Wait()
	}

	defer fc.mtx.Unlock()
	fc.flow -= 1
}
