package waitgroup

import (
	"sync/atomic"

	"primitives/internal/futex"
)

type WaitGroup struct {
	count uint32
}

func (wg *WaitGroup) Add(delta int) {
	v := int32(atomic.AddUint32(&wg.count, uint32(delta)))
	if v < 0 {
		panic("negative WaitGroup counter")
	}
	if v == 0 {
		futex.WakeAll(&wg.count)
	}
}

func (wg *WaitGroup) Done() {
	wg.Add(-1)
}

func (wg *WaitGroup) Wait() {
	for {
		v := atomic.LoadUint32(&wg.count)
		if v == 0 {
			return
		}
		futex.Wait(&wg.count, v)
	}
}
