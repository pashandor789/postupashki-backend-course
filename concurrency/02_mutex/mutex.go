package mutex

import (
	"primitives/internal/futex"
	"sync/atomic"
)

const (
	free      = iota // Мьютекс свободен
	held             // Мьютекс занят, никто не ждет
	contended        // Мьютекс занят, кто-то ждет
)

const maxWaitingIterations = 4

type Mutex struct {
	state uint32
}

func (m *Mutex) Lock() {
	// крутимся, ждем пока освободится TTAS если никто не ждет
	iterations := 0
	for iterations < maxWaitingIterations {
		if atomic.LoadUint32(&m.state) == free {
			if atomic.CompareAndSwapUint32(&m.state, free, held) {
				return
			}
		}
		iterations += 1
	}

	for {

		c := atomic.LoadUint32(&m.state)

		if c == contended || atomic.CompareAndSwapUint32(&m.state, held, contended) {
			futex.Wait(&m.state, contended)
		}

		if atomic.CompareAndSwapUint32(&m.state, free, contended) {
			return
		}
	}
}

func (m *Mutex) TryLock() bool {
	return atomic.CompareAndSwapUint32(&m.state, free, held)
}

func (m *Mutex) Unlock() {
	old := atomic.LoadUint32(&m.state)

	if old == free {
		panic("Mutex не используется")
	}

	new := atomic.AddUint32(&m.state, ^uint32(0))
	if new != free {
		atomic.SwapUint32(&m.state, free)
		futex.Wake(&m.state)
	}
}
