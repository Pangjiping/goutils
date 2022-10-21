package utils

import (
	"fmt"
	"log"
	"reflect"
	"runtime"
	"unsafe"
)

func Recovery() error {
	if err := recover(); err != nil {
		var e error
		switch r := err.(type) {
		case error:
			e = r
		default:
			e = fmt.Errorf("%v", r)
		}
		stack := make([]byte, 2048)
		length := runtime.Stack(stack, true)
		log.Printf("[PANIC RECOVER] %++v %++v\n", e, stack[:length])
		return e
	}
	return nil
}

// String2Bytes convert string to bytes.
func String2Bytes(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

// Bytes2String convert bytes to string.
func Bytes2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
