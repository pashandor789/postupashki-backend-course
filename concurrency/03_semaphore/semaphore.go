package semaphore

import (
	"sync/atomic"

	"primitives/internal/futex"
)

type Semaphore struct {
	permits uint32
}

func New(n int) *Semaphore {
	return &Semaphore{permits: uint32(n)}
}

func (s *Semaphore) Acquire() {
	for {
		if s.TryAcquire() {
			return
		}
		futex.Wait(&s.permits, 0)
	}
}

func (s *Semaphore) TryAcquire() bool {
	for {
		v := atomic.LoadUint32(&s.permits)
		if v == 0 {
			return false
		}
		if atomic.CompareAndSwapUint32(&s.permits, v, v-1) {
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
