package main

import "fmt"

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
	return vm.run()
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

		case OP_ADD:
			b := vm.pop()
			a := vm.pop()
			vm.push(a+b)
		case OP_SUBTRACT:
			b := vm.pop()
			a := vm.pop()
			vm.push(a-b)
		case OP_MULTIPLY:
			b := vm.pop()
			a := vm.pop()
			vm.push(a*b)
		case OP_DIVIDE:
			b := vm.pop()
			a := vm.pop()
			vm.push(a/b)

		case OP_NEGATE:
			vm.push(-vm.pop())

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
