package waitgroup

import (
	"primitives/internal/futex"
	"sync/atomic"
)

type WaitGroup struct {
	count uint32
}

func (wg *WaitGroup) Add(delta int) {
	atomic.AddUint32(&wg.count, uint32(delta))
}

func (wg *WaitGroup) Done() {
	new := atomic.AddUint32(&wg.count, ^uint32(0))

	if new == ^uint32(0) {
		panic("Счетчик не может быть меньше нуля")
	}
	if new == 0 {
		futex.WakeAll(&wg.count)
	}
}

func (wg *WaitGroup) Wait() {
	for {
		curr := atomic.LoadUint32(&wg.count)
		if curr == 0 {
			return
		}

		futex.Wait(&wg.count, curr)
	}
}
