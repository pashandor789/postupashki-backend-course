package semaphore

import (
	"primitives/internal/futex"
	"sync/atomic"
)

type Semaphore struct {
	permits uint32
}

func New(n int) *Semaphore {
	return &Semaphore{permits: uint32(n)}
}

func (s *Semaphore) Acquire() {
	for {
		current := atomic.LoadUint32(&s.permits)

		if current == 0 {
			futex.Wait(&s.permits, 0)
			continue
		}

		if atomic.CompareAndSwapUint32(&s.permits, current, current-1) {
			return
		}
	}
}

func (s *Semaphore) TryAcquire() bool {
	for {
		current := atomic.LoadUint32(&s.permits)

		if current == 0 {
			return false
		}

		if atomic.CompareAndSwapUint32(&s.permits, current, current-1) {
			return true
		}
	}
}

func (s *Semaphore) Release() {
	atomic.AddUint32(&s.permits, 1)

	futex.Wake(&s.permits)
}

func (s *Semaphore) Available() int {
	return int(atomic.LoadUint32(&s.permits))
}
