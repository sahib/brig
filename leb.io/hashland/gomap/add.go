package gomap

import "unsafe"

func add(p unsafe.Pointer, amt int) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + uintptr(amt))
}