package once

type Once struct {
	state uint32
}

func (o *Once) Do(f func()) {
	panic("не реализовано")
}

func (o *Once) Done() bool {
	panic("не реализовано")
}
