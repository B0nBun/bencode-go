package main

import (
	"runtime"
	"fmt"
	"reflect"
)

func assert(cond bool, format string, a ...any) {
	if !cond {
		_, filename, line, _ := runtime.Caller(1)
		msg := fmt.Sprintf(format, a...)
		panic(fmt.Sprintf("assertion error [%s:%d]: %s\n", filename, line, msg))
	}
}

func indirect(v reflect.Value) (rv reflect.Value, isNil bool) {
	for ; v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface; v = v.Elem() {
		if v.IsNil() {
			return v, true
		}
	}
	return v, false
}

func isAnyUint(k reflect.Kind) bool {
	return k == reflect.Uint || k == reflect.Uint8 || k == reflect.Uint16 || k == reflect.Uint32 || k == reflect.Uint64
}
