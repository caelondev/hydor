package main

import "fmt"

type ValueType int

const (
	VAL_BOOL ValueType = iota
	VAL_NIL
	VAL_NUMBER
)

type Value struct {
	Type   ValueType
	Number float64
	Bool   bool
}

type ValueArray struct {
	Values []Value
}

func NumberVal(n float64) Value {
	return Value{Type: VAL_NUMBER, Number: n}
}

func BoolVal(b bool) Value {
	return Value{Type: VAL_BOOL, Bool: b}
}

func NilVal() Value {
	return Value{Type: VAL_NIL}
}

func (v Value) IsBool() bool   { return v.Type == VAL_BOOL }
func (v Value) IsNil() bool    { return v.Type == VAL_NIL }
func (v Value) IsNumber() bool { return v.Type == VAL_NUMBER }
func (v Value) IsFalsy() bool {
	return v.IsNil() || (v.IsBool() && !v.AsBool())
}

func (v Value) AsBool() bool     { return v.Bool }
func (v Value) AsNumber() float64 { return v.Number }

func NewValueArray() *ValueArray {
	return &ValueArray{
		Values: make([]Value, 0),
	}
}

func (va *ValueArray) Write(value Value) {
	va.Values = append(va.Values, value)
}

func printValue(value Value) {
	switch value.Type {
	case VAL_BOOL:
		if value.Bool {
			fmt.Print("true")
		} else {
			fmt.Print("false")
		}
	case VAL_NIL:
		fmt.Print("nil")
	case VAL_NUMBER:
		fmt.Printf("%g", value.Number)
	}
}

func valuesEqual(a, b Value) bool {
	if a.Type != b.Type {
		return false
	}
	
	switch a.Type {
	case VAL_BOOL:
		return a.Bool == b.Bool
	case VAL_NIL:
		return true
	case VAL_NUMBER:
		return a.Number == b.Number
	default:
		return false
	}
}

func valueTypeName(t ValueType) string {
	switch t {
	case VAL_BOOL: return "boolean"
	case VAL_NIL: return "nil"
	case VAL_NUMBER: return "number"
	default: return "unknown"
	}
}
