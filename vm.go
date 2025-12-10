package main

const DEBUG_TRACE_EXECUTION = true

type InterpretResult int

const (
	INTERPRET_OK InterpretResult = iota
	INTERPRET_COMPILE_ERROR
	INTERPRET_RUNTIME_ERROR
)

type VM struct {
	Chunk *Chunk
	Ip int
}

func NewVM() *VM {
	return &VM{}
}

func (vm *VM) Interpret(chunk *Chunk) InterpretResult {
	vm.Chunk = chunk
	vm.Ip = 0

	return vm.run()
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
			vm.Chunk.DisassembleInstruction(vm.Ip)
		}

		instruction := readByte()
		switch OpCode(instruction) {
		case OP_CONSTANT:
			constant := readConstant()
			printValue(constant)
		case OP_RETURN: return INTERPRET_OK
		}
	}
}
