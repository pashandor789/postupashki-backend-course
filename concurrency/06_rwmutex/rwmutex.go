package rwmutex

import (
	"sync/atomic"

	"primitives/internal/futex"
)

const writer = 1 << 31

type RWMutex struct {
	state uint32
}

func (rw *RWMutex) RLock() {
	for {
		v := atomic.LoadUint32(&rw.state)
		if v&writer == 0 {
			if atomic.CompareAndSwapUint32(&rw.state, v, v+1) {
				return
			}
			continue
		}
		futex.Wait(&rw.state, v)
	}
}

func (rw *RWMutex) RUnlock() {
	for {
		v := atomic.LoadUint32(&rw.state)
		if v == 0 || v&writer != 0 {
			panic("RUnlock of unlocked RWMutex")
		}
		if atomic.CompareAndSwapUint32(&rw.state, v, v-1) {
			if v-1 == 0 {
				futex.WakeAll(&rw.state)
			}
			return
		}
	}
}

func (rw *RWMutex) Lock() {
	for {
		v := atomic.LoadUint32(&rw.state)
		if v == 0 && atomic.CompareAndSwapUint32(&rw.state, 0, writer) {
			return
		}
		if v != 0 {
			futex.Wait(&rw.state, v)
		}
	}
}

func (rw *RWMutex) Unlock() {
	if !atomic.CompareAndSwapUint32(&rw.state, writer, 0) {
		panic("Unlock of unlocked RWMutex")
	}
	futex.WakeAll(&rw.state)
}
