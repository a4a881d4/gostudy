package main

import (
	"fmt"
)

func main() {
	a := [...]int{1, 2, 3, 4}
	m, n := firstTwo(a[2:])
	fmt.Printf("first two item of a[2:] is %d,%d.\n", m, n)
}
func firstTwo(x []int) (int, int) {
	return x[0], x[1]
}
