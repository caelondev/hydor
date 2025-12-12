package debug

import (
	"fmt"

	"github.com/caelondev/hydor/frontend/bytecode"
	"github.com/caelondev/hydor/runtime/value"
)

func DisassembleBytecode(bc *bytecode.Bytecode, name string) {
	fmt.Printf("\t===== %s =====\n", name)

	offset := 0
	for offset < len(bc.Code) {
    offset = DisassembleInstruction(bc, offset);
  }
}

func DisassembleInstruction(bc *bytecode.Bytecode, offset int) int {
	line := GetLine(bc, offset)
	fmt.Printf("%04d ", offset)

	if offset > 0 && line == GetLine(bc, offset-1) {
		fmt.Printf("   ^ ")
	} else {
		fmt.Printf("%4d ", line)
	}

	instruction := bytecode.OpCode(bc.Code[offset])

	switch instruction {
	case bytecode.OP_CONSTANT: return constantInstruction("OP_CONSTANT", offset, bc)
	case bytecode.OP_NEGATE: return simpleInstruction("OP_NEGATE", offset)

	case bytecode.OP_TRUE: return simpleInstruction("OP_TRUE", offset)
	case bytecode.OP_FALSE: return simpleInstruction("OP_FALSE", offset)
	case bytecode.OP_NIL: return simpleInstruction("OP_NIL", offset)
	case bytecode.OP_NOT: return simpleInstruction("OP_NOT", offset)
	
	case bytecode.OP_EQUAL: return simpleInstruction("OP_EQUAL", offset)
	case bytecode.OP_LESS: return simpleInstruction("OP_LESS", offset)
	case bytecode.OP_GREATER: return simpleInstruction("OP_GREATER", offset)

	case bytecode.OP_ADD: return simpleInstruction("OP_ADD", offset)
	case bytecode.OP_SUBTRACT: return simpleInstruction("OP_SUBTRACT", offset)
	case bytecode.OP_MULTIPLY: return simpleInstruction("OP_MULTIPLY", offset)
	case bytecode.OP_DIVIDE: return simpleInstruction("OP_DIVIDE", offset)

	case bytecode.OP_RETURN: return simpleInstruction("OP_RETURN", offset)
	default:
		fmt.Printf("Unrecognized opcode '%d'\n", instruction)
		return move(offset, 1)
	}
}

func constantInstruction(name string, offset int, chunk *bytecode.Bytecode) int {
	constant := chunk.Code[offset+1]
	fmt.Printf("%-12s %4d  ", name, constant)
	value.PrintValue(chunk.Constants.Values[constant])
	fmt.Println()
	return move(offset, 2)
}

func simpleInstruction(name string, offset int) int {
	fmt.Printf("%-12s\n", name)
	return move(offset, 1)
}

func move(offset int, n int) int {
	return offset + n
}

func GetLine(bc *bytecode.Bytecode, offset int) int {
	instructionsSoFar := 0

	for _, run := range bc.Lines {
		if offset < instructionsSoFar+run.Count {
			return run.Line
		}
		instructionsSoFar += run.Count
	}

	return -1
}
