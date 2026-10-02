package rwmutex

import (
	"primitives/internal/futex"
	"sync/atomic"
)

const writer = 1 << 31

type RWMutex struct {
	state uint32
}

func (rw *RWMutex) RLock() { // Блокирует на чтение
	for {
		curr := atomic.LoadUint32(&rw.state)

		if (curr & writer) != 0 {
			futex.Wait(&rw.state, curr)
			continue
		}

		if atomic.CompareAndSwapUint32(&rw.state, curr, curr+1) {
			return // Успешно захватили блокировку на чтение
		}
	}
}

func (rw *RWMutex) RUnlock() { // Разблокирует на чтение
	new := atomic.AddUint32(&rw.state, ^uint32(0))

	if new == ^uint32(0) {
		panic("RWMutex не заблокирован")
	}

	futex.WakeAll(&rw.state)
}

func (rw *RWMutex) Lock() { // Блокирует на запись
	for {
		curr := atomic.LoadUint32(&rw.state)

		if curr == 0 {
			if atomic.CompareAndSwapUint32(&rw.state, 0, writer) {
				return
			}
			continue
		}

		futex.Wait(&rw.state, curr)
	}
}

func (rw *RWMutex) Unlock() { // Блокирует на запись
	if !atomic.CompareAndSwapUint32(&rw.state, writer, 0) {
		panic("RWMutex не заблокирован")
	}

	futex.WakeAll(&rw.state)
}
