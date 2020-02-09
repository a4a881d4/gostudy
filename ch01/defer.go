package main

import "fmt"

func testRecover(){
	defer func(){ 
		if p := recover(); p != nil{
			 fmt.Println(p) 
		}
	 }()
	panic("error happend")
}

func main() {
	testRecover()
	fmt.Println("I am alive")
}
