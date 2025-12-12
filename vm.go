package main

import (
	"fmt"
	"os"
)

const DEBUG_TRACE_EXECUTION = false
const STACK_MAX = 8 * 1024 * 1024

type InterpretResult int

const (
	INTERPRET_OK InterpretResult = iota
	INTERPRET_COMPILE_ERROR
	INTERPRET_RUNTIME_ERROR
)

type VM struct {
	Chunk    *Chunk
	Ip       int
	Stack    [STACK_MAX]Value
	StackTop int
}

func NewVM() *VM {
	return &VM{}
}

func (vm *VM) Interpret(chunk *Chunk) InterpretResult {

	vm.Chunk = chunk
	vm.Ip = 0
	vm.resetStack()

	result := vm.run()
	return result
}

func (vm *VM) resetStack() {
	vm.StackTop = 0
}

func (vm *VM) run() InterpretResult {
	readByte := func() byte {
		instruction := vm.Chunk.Code[vm.Ip]
		vm.Ip++
		return instruction
	}

	readConstant := func() Value {
		return vm.Chunk.Constants.Values[readByte()]
	}

	for {
		if DEBUG_TRACE_EXECUTION {
			for i := 0; i < vm.StackTop; i++ {
				fmt.Printf("[ ")
				printValue(vm.Stack[i])
				fmt.Printf(" ]")
				fmt.Println()
			}

			vm.Chunk.DisassembleInstruction(vm.Ip)
		}

		instruction := readByte()
		switch OpCode(instruction) {
		case OP_CONSTANT:
			constant := readConstant()
			vm.push(constant)

		case OP_NIL:
			vm.push(NilVal())
		case OP_TRUE:
			vm.push(BoolVal(true))
		case OP_FALSE:
			vm.push(BoolVal(false))
		case OP_NOT:
			vm.push(BoolVal(vm.pop().IsFalsy()))

		case OP_ADD:
			if vm.peek(0).IsString() && vm.peek(1).IsString() {
				vm.concatenate()
			} else if vm.peek(0).IsNumber() && vm.peek(1).IsNumber() {
				b := vm.pop()
				a := vm.pop()
				vm.push(NumberVal(a.AsNumber() + b.AsNumber()))
			} else {
				b := vm.pop()
				a := vm.pop()
				vm.runtimeError("Cannot add %s (%s) and %s (%s). Both operands must be numbers or both must be strings.",
					valueTypeName(a), formatValue(a),
					valueTypeName(b), formatValue(b))
				return INTERPRET_RUNTIME_ERROR
			}

		case OP_SUBTRACT:
			b := vm.pop()
			a := vm.pop()
			if !a.IsNumber() || !b.IsNumber() {
				vm.runtimeError("Cannot subtract %s (%s) from %s (%s). Both operands must be numbers.",
					valueTypeName(b), formatValue(b),
					valueTypeName(a), formatValue(a))
				return INTERPRET_RUNTIME_ERROR
			}
			vm.push(NumberVal(a.AsNumber() - b.AsNumber()))

		case OP_MULTIPLY:
			b := vm.pop()
			a := vm.pop()
			if !a.IsNumber() || !b.IsNumber() {
				vm.runtimeError("Cannot multiply %s (%s) by %s (%s). Both operands must be numbers.",
					valueTypeName(a), formatValue(a),
					valueTypeName(b), formatValue(b))
				return INTERPRET_RUNTIME_ERROR
			}
			vm.push(NumberVal(a.AsNumber() * b.AsNumber()))

		case OP_DIVIDE:
			b := vm.pop()
			a := vm.pop()
			if !a.IsNumber() || !b.IsNumber() {
				vm.runtimeError("Cannot divide %s (%s) by %s (%s). Both operands must be numbers.",
					valueTypeName(a), formatValue(a),
					valueTypeName(b), formatValue(b))
				return INTERPRET_RUNTIME_ERROR
			}

			if b.AsNumber() == 0 {
				vm.runtimeError("Cannot divide %g by zero. Division by zero is undefined.", a.AsNumber())
				return INTERPRET_RUNTIME_ERROR
			}
			vm.push(NumberVal(a.AsNumber() / b.AsNumber()))

		case OP_EQUAL:
			b := vm.pop()
			a := vm.pop()
			vm.push(BoolVal(valuesEqual(a, b)))
			
		case OP_GREATER:
			b := vm.pop()
			a := vm.pop()
			if !a.IsNumber() || !b.IsNumber() {
				vm.runtimeError("Cannot compare %s (%s) > %s (%s). Comparison operators require numeric operands.",
					valueTypeName(a), formatValue(a),
					valueTypeName(b), formatValue(b))
				return INTERPRET_RUNTIME_ERROR
			}
			vm.push(BoolVal(a.AsNumber() > b.AsNumber()))

		case OP_LESS:
			b := vm.pop()
			a := vm.pop()
			if !a.IsNumber() || !b.IsNumber() {
				vm.runtimeError("Cannot compare %s (%s) < %s (%s). Comparison operators require numeric operands.",
					valueTypeName(a), formatValue(a),
					valueTypeName(b), formatValue(b))
				return INTERPRET_RUNTIME_ERROR
			}
			vm.push(BoolVal(a.AsNumber() < b.AsNumber()))

		case OP_NEGATE:
			value := vm.pop()
			if !value.IsNumber() {
				vm.runtimeError("Cannot negate %s (%s). Unary '-' operator requires a numeric operand.",
					valueTypeName(value), formatValue(value))
				return INTERPRET_RUNTIME_ERROR
			}

			vm.push(NumberVal(-value.AsNumber()))

		case OP_RETURN:
			printValue(vm.pop())
			fmt.Println()
			return INTERPRET_OK
		}
	}
}

func (vm *VM) push(value Value) {
	vm.Stack[vm.StackTop] = value
	vm.StackTop++
}

func (vm *VM) pop() Value {
	vm.StackTop--
	return vm.Stack[vm.StackTop]
}

func (vm *VM) peek(distance int) Value {
	return vm.Stack[vm.StackTop-1-distance]
}

func (vm *VM) runtimeError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Runtime Error: ")
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintln(os.Stderr)

	instruction := vm.Ip - 1
	line := vm.Chunk.GetLine(instruction)
	fmt.Fprintf(os.Stderr, "    [line %d] in script\n", line)

	vm.resetStack()
}

func (vm *VM) concatenate() {
	b := vm.pop().AsString()
	a := vm.pop().AsString()

	str := NewString(a.Chars + b.Chars)
	vm.push(ObjVal(str.AsObj()))
}

// Helper function to format values for error messages
func formatValue(v Value) string {
	switch v.Type {
	case VAL_BOOL:
		if v.AsBool() {
			return "true"
		}
		return "false"
	case VAL_NIL:
		return "nil"
	case VAL_NUMBER:
		return fmt.Sprintf("%g", v.AsNumber())
	case VAL_OBJ:
		if v.IsString() {
			return fmt.Sprintf("\"%s\"", v.AsCString())
		}
		return "object"
	default:
		return "unknown"
	}
}
