package vm

import (
	"fmt"
	"math"
	"os"

	"github.com/caelondev/hydor/frontend/bytecode"
	"github.com/caelondev/hydor/frontend/debug"
	"github.com/caelondev/hydor/result"
	"github.com/caelondev/hydor/runtime/value"
)

const DEBUG_TRACE_EXECUTION = false
const STACK_MAX = 8 * 1024 * 1024

type VM struct {
	Bytecode *bytecode.Bytecode
	Ip       int
	Stack    [STACK_MAX]value.Value
	StackTop int
}

func NewVM() *VM {
	return &VM{}
}

func (vm *VM) Interpret(chunk *bytecode.Bytecode) result.InterpretResult {

	vm.Bytecode = chunk
	vm.Ip = 0
	vm.resetStack()

	result := vm.run()
	return result
}

func (vm *VM) resetStack() {
	vm.StackTop = 0
}

func (vm *VM) run() result.InterpretResult {
	readByte := func() byte {
		instruction := vm.Bytecode.Code[vm.Ip]
		vm.Ip++
		return instruction
	}

	readConstant := func() value.Value {
		return vm.Bytecode.Constants.Values[readByte()]
	}

	for {
		if DEBUG_TRACE_EXECUTION {
			for i := 0; i < vm.StackTop; i++ {
				fmt.Printf("[ ")
				value.PrintValue(vm.Stack[i])
				fmt.Printf(" ]")
				fmt.Println()
			}

			debug.DisassembleInstruction(vm.Bytecode, vm.Ip)
		}

		instruction := readByte()
		switch bytecode.OpCode(instruction) {
		case bytecode.OP_CONSTANT:
			constant := readConstant()
			vm.push(constant)

		case bytecode.OP_NIL:
			vm.push(value.NilVal())
		case bytecode.OP_TRUE:
			vm.push(value.BoolVal(true))
		case bytecode.OP_FALSE:
			vm.push(value.BoolVal(false))
		case bytecode.OP_NOT:
			vm.push(value.BoolVal(vm.pop().IsFalsy()))

		case bytecode.OP_ADD:
			if vm.peek(0).IsString() && vm.peek(1).IsString() {
				vm.concatenate()
			} else if vm.peek(0).IsNumber() && vm.peek(1).IsNumber() {
				b := vm.pop()
				a := vm.pop()
				vm.push(value.NumberVal(a.AsNumber() + b.AsNumber()))
			} else {
				b := vm.pop()
				a := vm.pop()
				vm.runtimeError("Cannot add %s (%s) and %s (%s). Both operands must be numbers or both must be strings.",
					value.ValueTypeName(a), formatValue(a),
					value.ValueTypeName(b), formatValue(b))
				return result.INTERPRET_RUNTIME_ERROR
			}

		case bytecode.OP_SUBTRACT:
			b := vm.pop()
			a := vm.pop()
			if !a.IsNumber() || !b.IsNumber() {
				vm.runtimeError("Cannot subtract %s (%s) from %s (%s). Both operands must be numbers.",
					value.ValueTypeName(b), formatValue(b),
					value.ValueTypeName(a), formatValue(a))
				return result.INTERPRET_RUNTIME_ERROR
			}
			vm.push(value.NumberVal(a.AsNumber() - b.AsNumber()))

		case bytecode.OP_MULTIPLY:
			b := vm.pop()
			a := vm.pop()
			if !a.IsNumber() || !b.IsNumber() {
				vm.runtimeError("Cannot multiply %s (%s) by %s (%s). Both operands must be numbers.",
					value.ValueTypeName(a), formatValue(a),
					value.ValueTypeName(b), formatValue(b))
				return result.INTERPRET_RUNTIME_ERROR
			}
			vm.push(value.NumberVal(a.AsNumber() * b.AsNumber()))

		case bytecode.OP_DIVIDE:
			b := vm.pop()
			a := vm.pop()
			if !a.IsNumber() || !b.IsNumber() {
				vm.runtimeError("Cannot divide %s (%s) by %s (%s). Both operands must be numbers.",
					value.ValueTypeName(a), formatValue(a),
					value.ValueTypeName(b), formatValue(b))
				return result.INTERPRET_RUNTIME_ERROR
			}

			if b.AsNumber() == 0 {
				vm.runtimeError("Cannot divide %g by zero. Division by zero is undefined.", a.AsNumber())
				return result.INTERPRET_RUNTIME_ERROR
			}
			vm.push(value.NumberVal(a.AsNumber() / b.AsNumber()))

		case bytecode.OP_MODULO:
			b := vm.pop()
			a := vm.pop()
			if !a.IsNumber() || !b.IsNumber() {
				vm.runtimeError("Cannot divide %s (%s) by %s (%s). Both operands must be numbers.",
					value.ValueTypeName(a), formatValue(a),
					value.ValueTypeName(b), formatValue(b))
				return result.INTERPRET_RUNTIME_ERROR
			}

			if b.AsNumber() == 0 {
				vm.runtimeError("Cannot modulo %g by zero. Division by zero is undefined.", a.AsNumber())
				return result.INTERPRET_RUNTIME_ERROR
			}
		 vm.push(value.NumberVal(math.Mod(a.AsNumber(), b.AsNumber())))

		case bytecode.OP_EQUAL:
			b := vm.pop()
			a := vm.pop()
			vm.push(value.BoolVal(value.ValuesEqual(a, b)))

		case bytecode.OP_GREATER:
			b := vm.pop()
			a := vm.pop()
			if !a.IsNumber() || !b.IsNumber() {
				vm.runtimeError("Cannot compare %s (%s) > %s (%s). Comparison operators require numeric operands.",
					value.ValueTypeName(a), formatValue(a),
					value.ValueTypeName(b), formatValue(b))
				return result.INTERPRET_RUNTIME_ERROR
			}
			vm.push(value.BoolVal(a.AsNumber() > b.AsNumber()))

		case bytecode.OP_LESS:
			b := vm.pop()
			a := vm.pop()
			if !a.IsNumber() || !b.IsNumber() {
				vm.runtimeError("Cannot compare %s (%s) < %s (%s). Comparison operators require numeric operands.",
					value.ValueTypeName(a), formatValue(a),
					value.ValueTypeName(b), formatValue(b))
				return result.INTERPRET_RUNTIME_ERROR
			}
			vm.push(value.BoolVal(a.AsNumber() < b.AsNumber()))

		case bytecode.OP_NEGATE:
			val := vm.pop()
			if !val.IsNumber() {
				vm.runtimeError("Cannot negate %s (%s). Unary '-' operator requires a numeric operand.",
					value.ValueTypeName(val), formatValue(val))
				return result.INTERPRET_RUNTIME_ERROR
			}

			vm.push(value.NumberVal(-val.AsNumber()))

		case bytecode.OP_RETURN:
			value.PrintValue(vm.pop())
			fmt.Println()
			return result.INTERPRET_OK
		}
	}
}

func (vm *VM) push(value value.Value) {
	vm.Stack[vm.StackTop] = value
	vm.StackTop++
}

func (vm *VM) pop() value.Value {
	vm.StackTop--
	return vm.Stack[vm.StackTop]
}

func (vm *VM) peek(distance int) value.Value {
	return vm.Stack[vm.StackTop-1-distance]
}

func (vm *VM) runtimeError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Runtime Error: ")
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintln(os.Stderr)

	instruction := vm.Ip - 1
	line := debug.GetLine(vm.Bytecode, instruction)
	fmt.Fprintf(os.Stderr, "    [line %d] in script\n", line)

	vm.resetStack()
}

func (vm *VM) concatenate() {
	b := vm.pop().AsString()
	a := vm.pop().AsString()

	str := value.NewString(a.Chars + b.Chars)
	vm.push(value.ObjVal(str.AsObj()))
}

func formatValue(v value.Value) string {
	switch v.Type {
	case value.VAL_BOOL:
		if v.AsBool() {
			return "true"
		}
		return "false"
	case value.VAL_NIL:
		return "nil"
	case value.VAL_NUMBER:
		return fmt.Sprintf("%g", v.AsNumber())
	case value.VAL_OBJ:
		if v.IsString() {
			return fmt.Sprintf("\"%s\"", v.AsCString())
		}
		return "object"
	default:
		return "unknown"
	}
}
