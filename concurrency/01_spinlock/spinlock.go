package spinlock

import "sync/atomic"

type Spinlock struct {
	locked atomic.Bool
}

func (s *Spinlock) Lock() {
	for !s.locked.CompareAndSwap(false, true) {
	}
}

func (s *Spinlock) TryLock() bool {
	return s.locked.CompareAndSwap(false, true)
}

func (s *Spinlock) Unlock() {
	if !s.locked.Load() {
		panic("Spinlock не захвачен")
	}
	s.locked.Swap(false)
}

type TTAS struct {
	locked atomic.Bool
}

func (s *TTAS) Lock() {
	locked := true
	for locked {
		locked = s.locked.Load()

		if !locked {
			locked = !s.locked.CompareAndSwap(false, true)
		}
	}
}

func (s *TTAS) TryLock() bool {
	locked := s.locked.Load()
	if !locked {
		return s.locked.CompareAndSwap(false, true)
	}
	return false
}

func (s *TTAS) Unlock() {
	if !s.locked.Load() {
		panic("Spinlock не захвачен")
	}
	s.locked.Swap(false)
}
