package main

type OpCode byte

const (
	OP_CONSTANT OpCode = iota
	OP_RETURN
)

type Chunk struct {
	Code []byte
	Lines []int
	Constants ValueArray
}

func NewChunk() *Chunk {
	return &Chunk{
		Code: make([]byte, 0),
		Lines: make([]int, 0),
		Constants: *NewValueArray(),
	}
}

func (c *Chunk) AddConstant(value Value) int {
	c.Constants.Write(value)

	return len(c.Constants.Values) -1
}

func (c *Chunk) Write(instruction byte, line int) {
	c.Code = append(c.Code, instruction)

	// RLE Compression ---
	if len(c.Lines) >= 2 && c.Lines[len(c.Lines)-2] == line {
		c.Lines[len(c.Lines)-1]++
	} else {
		c.Lines = append(c.Lines, line, 1)
	}
}


