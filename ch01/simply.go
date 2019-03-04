package main

import (
	"fmt"
)

func main() {
	a := [0x10000000]int{1, 2, 3, 4}
	m, n := firstTwo(a)
	fmt.Printf("first two item of a[2:] is %d,%d.\n", m, n)
	fmt.Println(a[:10])
}
func firstTwo(x [0x10000000]int) (int, int) {
	x[0]=5
	return x[0], x[1]
}
