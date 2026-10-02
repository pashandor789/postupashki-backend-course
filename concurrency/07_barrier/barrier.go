package barrier

import (
	"primitives/internal/futex"
	"sync/atomic"
)

type Barrier struct {
	need    uint32
	arrived uint32
	round   uint32
}

func New(n int) *Barrier {
	return &Barrier{need: uint32(n), arrived: 0, round: 0}
}

func (b *Barrier) Wait() {
	for {
		curr := atomic.LoadUint32(&b.arrived)
		round := atomic.LoadUint32(&b.round)

		if atomic.CompareAndSwapUint32(&b.arrived, curr, curr+1) {
			need := atomic.LoadUint32(&b.need)
			if curr+1 == need {
				atomic.AddUint32(&b.round, 1)

				atomic.SwapUint32(&b.arrived, 0)

				futex.WakeAll(&b.round)
				return
			}

			if round == atomic.LoadUint32(&b.round) {
				futex.Wait(&b.round, round)
			}
			return
		}
	}
}
