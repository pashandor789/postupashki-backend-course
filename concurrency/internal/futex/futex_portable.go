//go:build !linux

package futex

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

var (
	mu      sync.Mutex
	waiters = map[uintptr][]chan struct{}{}
)

func key(addr *uint32) uintptr {
	return uintptr(unsafe.Pointer(addr))
}

func Wait(addr *uint32, val uint32) {
	k := key(addr)
	ch := make(chan struct{}, 1)

	mu.Lock()
	if atomic.LoadUint32(addr) != val {
		mu.Unlock()
		return
	}
	waiters[k] = append(waiters[k], ch)
	mu.Unlock()

	<-ch
}

func Wake(addr *uint32) {
	k := key(addr)

	mu.Lock()
	list := waiters[k]
	if len(list) == 0 {
		mu.Unlock()
		return
	}
	ch := list[0]
	if len(list) == 1 {
		delete(waiters, k)
	} else {
		waiters[k] = list[1:]
	}
	mu.Unlock()

	ch <- struct{}{}
}

func WakeAll(addr *uint32) {
	k := key(addr)

	mu.Lock()
	list := waiters[k]
	delete(waiters, k)
	mu.Unlock()

	for _, ch := range list {
		ch <- struct{}{}
	}
}
