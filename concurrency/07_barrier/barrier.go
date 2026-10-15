package barrier

import (
	"sync/atomic"

	"primitives/internal/futex"
)

type Barrier struct {
	need    uint32
	arrived uint32
	round   uint32
}

func New(n int) *Barrier {
	return &Barrier{need: uint32(n)}
}

func (b *Barrier) Wait() {
	r := atomic.LoadUint32(&b.round)
	if atomic.AddUint32(&b.arrived, 1) == b.need {
		atomic.StoreUint32(&b.arrived, 0)
		atomic.AddUint32(&b.round, 1)
		futex.WakeAll(&b.round)
		return
	}
	for atomic.LoadUint32(&b.round) == r {
		futex.Wait(&b.round, r)
	}
}
