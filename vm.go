package main

import (
	"fmt"
	"os"
)

const DEBUG_TRACE_EXECUTION = true
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

func (vm *VM) Interpret(source string) InterpretResult {
	chunk := NewChunk(source)
	parser := NewParser()
	
	if !parser.Compile(source, chunk) {
		return INTERPRET_COMPILE_ERROR
	}

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

		case OP_NIL: vm.push(NilVal())
		case OP_TRUE: vm.push(BoolVal(true))
		case OP_FALSE: vm.push(BoolVal(false))
		case OP_NOT: vm.push(BoolVal(vm.pop().IsFalsy()))

		// TODO : Should I rewrite this? or at least make a helper ---
		// function? this kinda looks ugly ---
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
				vm.runtimeError("Could not add nor concatenate operands (%s and %s)", valueTypeName(a), valueTypeName(b))
				return INTERPRET_RUNTIME_ERROR
			}

		case OP_SUBTRACT:
			if !vm.peek(0).IsNumber() || !vm.peek(1).IsNumber() {
				b := vm.pop()
				a := vm.pop()
				vm.runtimeError("Cannot subtract %s type to a %s type", valueTypeName(a), valueTypeName(b))
				return INTERPRET_RUNTIME_ERROR
			}
		case OP_MULTIPLY:
			if !vm.peek(0).IsNumber() || !vm.peek(1).IsNumber() {
				b := vm.pop()
				a := vm.pop()
				vm.runtimeError("Cannot multiply %s type to a %s type", valueTypeName(a), valueTypeName(b))
				return INTERPRET_RUNTIME_ERROR
			}
		case OP_DIVIDE:
			if !vm.peek(0).IsNumber() || !vm.peek(1).IsNumber() {
				b := vm.pop()
				a := vm.pop()
				vm.runtimeError("Cannot divide %s type to a %s type", valueTypeName(a), valueTypeName(b))
				return INTERPRET_RUNTIME_ERROR
			}

		case OP_EQUAL:
			b := vm.pop()
			a := vm.pop()
			vm.push(BoolVal(valuesEqual(a, b)))
		case OP_GREATER: 
			b := vm.pop()
			a := vm.pop()
			if !a.IsNumber() || !b.IsNumber() {
				vm.runtimeError("Could not perform relational operation on given value type (%s and %s)", valueTypeName(a), valueTypeName(b))

				return INTERPRET_RUNTIME_ERROR
			}

		case OP_LESS: 
			b := vm.pop()
			a := vm.pop()
			if !a.IsNumber() || !b.IsNumber() {
				vm.runtimeError("Could not perform relational operation on given value type (%s and %s)", valueTypeName(a), valueTypeName(b))
				return INTERPRET_RUNTIME_ERROR
			}

		case OP_NEGATE:
			value := vm.pop()
			if !value.IsNumber() {
				vm.runtimeError("Cannot use Logical NOT on a %s as it must be a boolean", valueTypeName(value))
				return INTERPRET_RUNTIME_ERROR
			}

			vm.push(BoolVal(!value.AsBool()))

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
	return vm.Stack[vm.StackTop - 1 - distance]
}

func (vm *VM) runtimeError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintln(os.Stderr)
	
	instruction := vm.Ip - 1
	line := vm.Chunk.GetLine(instruction)
	fmt.Fprintf(os.Stderr, "[line %d] in script\n", line)
	
	vm.resetStack()
}

func (vm *VM) concatenate() {
	b := vm.pop().AsString()
	a := vm.pop().AsString()

	str := NewString(a.Chars + b.Chars)
	vm.push(ObjVal(str.AsObj()))
}
