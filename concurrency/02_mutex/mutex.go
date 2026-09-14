package mutex

const (
	free = iota
	held
	contended
)

type Mutex struct {
	state uint32
}

func (m *Mutex) Lock() {
	panic("не реализовано")
}

func (m *Mutex) TryLock() bool {
	panic("не реализовано")
}

func (m *Mutex) Unlock() {
	panic("не реализовано")
}
