package spinlock

import "sync/atomic"

type Spinlock struct {
	locked atomic.Bool
}

func (s *Spinlock) Lock() {
	panic("не реализовано")
}

func (s *Spinlock) TryLock() bool {
	panic("не реализовано")
}

func (s *Spinlock) Unlock() {
	panic("не реализовано")
}

type TTAS struct {
	locked atomic.Bool
}

func (s *TTAS) Lock() {
	panic("не реализовано")
}

func (s *TTAS) TryLock() bool {
	panic("не реализовано")
}

func (s *TTAS) Unlock() {
	panic("не реализовано")
}
