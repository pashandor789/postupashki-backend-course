package rwmutex

const writer = 1 << 31

type RWMutex struct {
	state uint32
}

func (rw *RWMutex) RLock() {
	panic("не реализовано")
}

func (rw *RWMutex) RUnlock() {
	panic("не реализовано")
}

func (rw *RWMutex) Lock() {
	panic("не реализовано")
}

func (rw *RWMutex) Unlock() {
	panic("не реализовано")
}
