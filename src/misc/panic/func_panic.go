package main

// 使用延迟匿名函数或下面这样都是有效的

func except() {
	recover()
}

func test() {
	defer except()
	panic("test panic")
}

func main() {
	test()
}
