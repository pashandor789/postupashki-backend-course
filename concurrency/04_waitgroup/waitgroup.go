package waitgroup

type WaitGroup struct {
	count uint32
}

func (wg *WaitGroup) Add(delta int) {
	panic("не реализовано")
}

func (wg *WaitGroup) Done() {
	panic("не реализовано")
}

func (wg *WaitGroup) Wait() {
	panic("не реализовано")
}
