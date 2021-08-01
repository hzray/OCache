package main

import (
	"fmt"
	"time"
)

func main() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("goroutine recovered!")
			}
		}()
		time.Sleep(1 * time.Second)
		panic("goroutine panic")
	}()
	for {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("main")
	}
}
