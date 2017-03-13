package main

import (
	"fmt"
	"time"
)

func main() {
	s := spin.New()
	for i := 0; i < 30; i++ {
		fmt.Printf("\r  \033[36mcomputing\033[m %s ", s.Next())
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println()
}
