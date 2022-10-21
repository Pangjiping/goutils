package main

import "fmt"

func main() {
	a := []int{1, 2, 3}
	fmt.Printf("a's len: %v cap: %v data: %v\n", len(a), cap(a), a)
	fmt.Printf("a's addr: %p\n", a)
	app(a)
	fmt.Printf("a's len: %v cap: %v data: %v\n", len(a), cap(a), a)
	fmt.Printf("a's addr: %p\n", a)
}

func app(a []int) {
	a[0] = 100
	a = append(a, 4)
	fmt.Printf("app(a)'s len: %v cap: %v data: %v\n", len(a), cap(a), a)
	fmt.Printf("app(a)'s addr: %p\n", a)
}
