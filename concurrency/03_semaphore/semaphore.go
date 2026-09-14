package semaphore

type Semaphore struct {
	permits uint32
}

func New(n int) *Semaphore {
	panic("не реализовано")
}

func (s *Semaphore) Acquire() {
	panic("не реализовано")
}

func (s *Semaphore) TryAcquire() bool {
	panic("не реализовано")
}

func (s *Semaphore) Release() {
	panic("не реализовано")
}

func (s *Semaphore) Available() int {
	panic("не реализовано")
}
