package main

import (
	"fmt"
	"unsafe"
)

type people struct{
	Age int64
	State bool
}
func main() {
	fmt.Println(unsafe.Alignof(int64(2)))
	fmt.Println(unsafe.Alignof(true))
	fmt.Println(unsafe.Alignof(people{}.State))
	fmt.Println(unsafe.Alignof(people{}.Age))
}
