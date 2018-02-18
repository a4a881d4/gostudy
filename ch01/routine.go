package main

import (
	"fmt"
	"time"
)

func say(s string, order chan int) {
	for i := 0; ; i++ {
		o := <-order
		if o == 1 {
			fmt.Printf("%d: ", i)
			fmt.Println(s)
		} else {
			fmt.Println("Bye")
			break
		}
	}
}

func main() {
	order := make(chan int)
	go say("hello world", order)
	for i := 0; i < 10; i++ {
		time.Sleep(200 * time.Millisecond)
		order <- 1
	}
	order <- 0
}
