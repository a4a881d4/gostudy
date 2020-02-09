package main

import "fmt"
func swap1(a,b int){
	a,b = b,a
}
func swap2(a,b *int){
	*a,*b = *b,*a
}

func main() {
	a,b := 1,2
	swap1(a,b)
	fmt.Println(a,b)
	swap2(&a,&b)
	fmt.Println(a,b)
}
