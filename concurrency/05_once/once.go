package once

import (
	"primitives/internal/futex"
	"sync/atomic"
)

const (
	free = iota
	executing
	executed
)

type Once struct {
	state uint32
}

func (o *Once) Do(f func()) {
	if atomic.LoadUint32(&o.state) == executed {
		return
	}

	if atomic.CompareAndSwapUint32(&o.state, free, executing) {
		defer func() {
			atomic.StoreUint32(&o.state, executed)
			futex.WakeAll(&o.state)
		}()

		f()
		return
	}

	for {
		currentState := atomic.LoadUint32(&o.state)

		// Если пока мы собирались спать, поток А уже всё закончил — выходим
		if currentState == executed {
			return
		}

		// Засыпаем только если состояние до сих пор executing (1)
		futex.Wait(&o.state, currentState)
	}
}

func (o *Once) Done() bool {
	return atomic.LoadUint32(&o.state) == executed
}
