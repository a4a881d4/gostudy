package main

import (
	"log"
)

func protect(g func()) {
	defer func() {
		log.Println("Done")
		if x := recover(); x != nil {
			log.Printf("Run time panic: %v", x)
		}
	}()
	log.Println("Start")
	g()
}
func main() {
	var s []byte
	protect(func() { s[0] = 0 })
	protect(func() { panic(42) })
	s[0] = 10
}
