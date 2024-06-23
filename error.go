package main

import "reflect"

type InvalidUnmarshalError struct {
	Pos int
}

func (e InvalidUnmarshalError) Error() string {
	return "bencode"
}

type InvalidMarshalError struct {
	Val reflect.Value
}

func (e InvalidMarshalError) Error() string {
	return "bencode"
}

type SyntaxError struct{}

func (_ SyntaxError) Error() string {
	return "syntax error"
}

type ValueError struct{}

func (_ ValueError) Error() string {
	return "value error"
}