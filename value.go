package main

import (
	"fmt"
	"unsafe"
)

type ValueType int

const (
	VAL_BOOL ValueType = iota
	VAL_NIL
	VAL_NUMBER
	VAL_OBJ
)

type ObjType int

const (
	OBJ_STRING ObjType = iota
)

type Obj struct {
	Type ObjType
	Next *Obj
}

type ObjString struct {
	Object Obj
	Chars  string
	Length int
}

type Value struct {
	Type   ValueType
	Number float64
	Bool   bool
	Obj    *Obj
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

func ObjVal(obj *Obj) Value {
	return Value{Type: VAL_OBJ, Obj: obj}
}

func (v Value) IsBool() bool   { return v.Type == VAL_BOOL }
func (v Value) IsNil() bool    { return v.Type == VAL_NIL }
func (v Value) IsNumber() bool { return v.Type == VAL_NUMBER }
func (v Value) IsObj() bool    { return v.Type == VAL_OBJ }
func (v Value) IsString() bool { return v.IsObj() && v.Obj.Type == OBJ_STRING }

func (v Value) IsFalsy() bool {
	return v.IsNil() || (v.IsBool() && !v.AsBool())
}

func (v Value) AsBool() bool     { return v.Bool }
func (v Value) AsNumber() float64 { return v.Number }
func (v Value) AsObj() *Obj      { return v.Obj }
func (v Value) AsString() *ObjString {
	return (*ObjString)(unsafe.Pointer(v.Obj))
}
func (v Value) AsCString() string {
	return v.AsString().Chars
}

func NewString(chars string) *ObjString {
	str := &ObjString{
		Chars:  chars,
		Length: len(chars),
	}
	str.Object.Type = OBJ_STRING
	return str
}

func (s *ObjString) AsObj() *Obj {
	return &s.Object
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
	case VAL_OBJ:
		printObject(value)
	}
}

func printObject(value Value) {
	switch value.AsObj().Type {
	case OBJ_STRING:
		fmt.Print(value.AsCString())
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
	case VAL_OBJ:
		return objectsEqual(a, b)
	default:
		return false
	}
}

func objectsEqual(a, b Value) bool {
	if a.AsObj().Type != b.AsObj().Type {
		return false
	}

	switch a.AsObj().Type {
	case OBJ_STRING:
		return a.AsCString() == b.AsCString()
	default:
		return false
	}
}

func valueTypeName(v Value) string {
	switch v.Type {
	case VAL_BOOL:
		return "boolean"
	case VAL_NIL:
		return "nil"
	case VAL_NUMBER:
		return "number"
	case VAL_OBJ:
		switch v.AsObj().Type {
		case OBJ_STRING:
			return "string"
		default:
			return "object"
		}
	default:
		return "unknown"
	}
}
