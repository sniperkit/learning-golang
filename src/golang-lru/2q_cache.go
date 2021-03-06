package main

import "fmt"

func main() {
	l, err := lru.New2Q(10)
	if err != nil {
		return
	}

	for i := 0; i < 20; i++ {
		l.Add(i, i)
	}

	if l.Len() != 10 {
		panic(fmt.Sprintf("bad len: %v", l.Len()))
	}

	for i := 0; i < 20; i++ {
		fmt.Println(l.Get(i))
	}
}
