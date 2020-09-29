package main

import (
	"fmt"
	"time"
)

func main () {
	s := []int{}

	go func() {
		for i := 0; i < 9999999; i++ {
			s = append(s, i)
		}
	}()

	go func() {
		for idx, item := range(s) {
			fmt.Println(idx, item)
		}
	}()

	time.Sleep(1e9)

}
