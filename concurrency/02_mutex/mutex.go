package mutex

import (
	"sync/atomic"

	"primitives/internal/futex"
)

const (
	free = iota
	held
	contended
)

type Mutex struct {
	state uint32
}

func (m *Mutex) Lock() {
	if atomic.CompareAndSwapUint32(&m.state, free, held) {
		return
	}
	for atomic.SwapUint32(&m.state, contended) != free {
		futex.Wait(&m.state, contended)
	}
}

func (m *Mutex) TryLock() bool {
	return atomic.CompareAndSwapUint32(&m.state, free, held)
}

func (m *Mutex) Unlock() {
	switch atomic.SwapUint32(&m.state, free) {
	case free:
		panic("unlock of unlocked mutex")
	case contended:
		futex.Wake(&m.state)
	}
}
