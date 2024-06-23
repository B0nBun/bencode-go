package main

import (
	"reflect"
	"testing"
)

func testBasic[T any](t *testing.T, input string, expect T) {
	t.Helper()
	var got T
	err := Unmarshal([]byte(input), &got)
	if err != nil {
		t.Errorf("got unexpected error from input '%s': %v", input, err)
		return
	}
	deepEqual := reflect.DeepEqual(got, expect)
	ev := reflect.ValueOf(expect)
	gv := reflect.ValueOf(got)
	if ev.Kind() == reflect.Slice && gv.Len() == 0 && ev.Len() == 0 { // because nil slices and slices with no elements aren't equal
		deepEqual = true
	}
	if !deepEqual {
		t.Errorf("got %v (%v), wanted %v (%v) from input '%s'", got, reflect.TypeOf(got), expect, reflect.TypeOf(expect), input)
	}
}

func TestBasic(t *testing.T) {
	testBasic(t, "i42e", int64(42))
	testBasic(t, "i0e", uint(0))
	testBasic(t, "i-13e", int8(-13))

	testBasic(t, "4:spam", "spam")
	testBasic(t, "0:", "")

	testBasic(t, "li1ei2ei3ee", []int{1, 2, 3})
	testBasic(t, "li-2ei-1ee", [2]int{-2, -1})
	testBasic(t, "le", []int{})
	testBasic(t, "le", []string{})

	type Bar struct {
		L []int `bencode:"l"`
	}

	type Foo struct {
		N int    `bencode:"n"`
		S string `bencode:"?s"`
		D Bar    `bencode:"?d"`
	}

	foo := Foo{N: 42, S: "string", D: Bar{L: []int{1, 2}}}
	testBasic(t, "d1:dd1:lli1ei2eee1:ni42e1:s6:stringe", foo)

	foo = Foo{N: 42}
	testBasic(t, "d1:ni42ee", foo)
}
