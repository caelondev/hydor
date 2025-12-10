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
	fmt.Printf("%04d ", offset)

	if offset > 0 && c.GetLine(offset) == c.GetLine(offset - 1) {
		fmt.Printf("   ^ ")
	} else {
		fmt.Printf("%4d ", c.Lines[offset])
	}

	instruction := OpCode(c.Code[offset])

	switch instruction {
	case OP_CONSTANT: return constantInstruction("OP_CONSTANT", offset, c)
	case OP_NEGATE: return simpleInstruction("OP_NEGATE", offset)

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
	constant := chunk.Code[offset+1] // Takes the value after the constant instruction
	fmt.Printf("%-12s %4d", name, constant)
	printValue(chunk.Constants.Values[constant])

	fmt.Println()

	return move(offset, 2) // Advance past both constant and value
}

func simpleInstruction(name string, offset int) int {
	fmt.Printf("%s\n", name)
	return move(offset, 1)
}

func move(offset int, n int) int {
	return offset + n
}

func (c *Chunk) GetLine(offset int) int {
	instructionsSoFar := 0
	for i := 0; i < len(c.Lines); i += 2 {
		line := c.Lines[i]
		count := c.Lines[i+1]

		if offset < instructionsSoFar + count {
			return line
		}

		instructionsSoFar += count
	}

	return -1
}
