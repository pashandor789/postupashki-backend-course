//go:build linux

package futex

import (
	"math"
	"syscall"
	"unsafe"
)

const (
	opWait  = 0
	opWake  = 1
	private = 128
)

func Wait(addr *uint32, val uint32) {
	syscall.Syscall6(syscall.SYS_FUTEX, uintptr(unsafe.Pointer(addr)),
		opWait|private, uintptr(val), 0, 0, 0)
}

func Wake(addr *uint32) {
	syscall.Syscall6(syscall.SYS_FUTEX, uintptr(unsafe.Pointer(addr)),
		opWake|private, 1, 0, 0, 0)
}

func WakeAll(addr *uint32) {
	syscall.Syscall6(syscall.SYS_FUTEX, uintptr(unsafe.Pointer(addr)),
		opWake|private, uintptr(math.MaxInt32), 0, 0, 0)
}
