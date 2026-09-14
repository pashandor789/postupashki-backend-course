package barrier

type Barrier struct {
	need    uint32
	arrived uint32
	round   uint32
}

func New(n int) *Barrier {
	panic("не реализовано")
}

func (b *Barrier) Wait() {
	panic("не реализовано")
}
