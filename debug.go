package main

import "fmt"

func (c *Chunk) DisassembleChunk(name string) {
	fmt.Printf("\t===== %s =====\n", name)

	offset := 0
	for offset < len(c.Code) {
    offset = c.DisassembleInstruction(offset);
  }
}

func (c *Chunk) DisassembleInstruction(offset int) int {
	line := c.GetLine(offset)
	fmt.Printf("%04d ", offset)

	if offset > 0 && line == c.GetLine(offset-1) {
		fmt.Printf("   ^ ")
	} else {
		fmt.Printf("%4d ", line)
	}

	instruction := OpCode(c.Code[offset])

	switch instruction {
	case OP_CONSTANT: return constantInstruction("OP_CONSTANT", offset, c)
	case OP_NEGATE: return simpleInstruction("OP_NEGATE", offset)

	case OP_TRUE: return simpleInstruction("OP_TRUE", offset)
	case OP_FALSE: return simpleInstruction("OP_FALSE", offset)
	case OP_NIL: return simpleInstruction("OP_NIL", offset)
	case OP_NOT: return simpleInstruction("OP_NOT", offset)
	
	case OP_EQUAL: return simpleInstruction("OP_EQUAL", offset)
	case OP_LESS: return simpleInstruction("OP_LESS", offset)
	case OP_GREATER: return simpleInstruction("OP_GREATER", offset)

	case OP_ADD: return simpleInstruction("OP_ADD", offset)
	case OP_SUBTRACT: return simpleInstruction("OP_SUBTRACT", offset)
	case OP_MULTIPLY: return simpleInstruction("OP_MULTIPLY", offset)
	case OP_DIVIDE: return simpleInstruction("OP_DIVIDE", offset)

	case OP_RETURN: return simpleInstruction("OP_RETURN", offset)
	default:
		fmt.Printf("Unrecognized opcode '%d'\n", instruction)
		return move(offset, 1)
	}
}

func constantInstruction(name string, offset int, chunk *Chunk) int {
	constant := chunk.Code[offset+1]
	fmt.Printf("%-12s %4d  ", name, constant)
	printValue(chunk.Constants.Values[constant])
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

func (c *Chunk) GetLine(offset int) int {
	instructionsSoFar := 0

	for _, run := range c.Lines {
		if offset < instructionsSoFar+run.Count {
			return run.Line
		}
		instructionsSoFar += run.Count
	}

	return -1
}
