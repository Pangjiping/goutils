package utils

import (
	"fmt"
	"log"
	"runtime"
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
