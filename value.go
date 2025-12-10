package main

import "fmt"

type Value float64

type ValueArray struct {
	Values []Value
}

func NewValueArray() *ValueArray {
	return &ValueArray{
		Values: make([]Value, 0),
	}
}

func (va *ValueArray) Write(value Value) {
	va.Values = append(va.Values, value)
}

func printValue(value Value) {
	fmt.Printf("%g\n", value)
}
