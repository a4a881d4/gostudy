package main

import (
	"fmt"
	"unsafe"
)

type W struct {
	a byte
	b int
	c int64
}

func main() {
	w := new(W)
	s := (int)(unsafe.Sizeof(*w))
	fmt.Println("Size of W is ", s)
	w.a = 1
	w.b = 2
	w.c = 3

	for i := 0; i < s; i++ {
		p := (*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(w)) + uintptr(i)))
		fmt.Printf("%02x", *p)
	}
	fmt.Println("")
}
