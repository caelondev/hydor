package main

type OpCode byte
const (
	OP_CONSTANT OpCode = iota
	OP_ADD
	OP_SUBTRACT
	OP_MULTIPLY
	OP_DIVIDE
	OP_NEGATE
	OP_RETURN
)

type Chunk struct {
	Code      []byte
	Lines     []LineRun
	Constants ValueArray
}

type LineRun struct { Line, Count int }

func NewChunk(src string) *Chunk {
	return &Chunk{
		Code:      make([]byte, 0, len(src)+16),
		Lines:     make([]LineRun, 0, len(src)/4),
		Constants: *NewValueArray(),
	}
}

func (c *Chunk) AddConstant(v Value) int {
	c.Constants.Write(v)
	return len(c.Constants.Values) - 1
}

func (c *Chunk) Write(b byte, line int) {
	c.Code = append(c.Code, b)
	if len(c.Lines) > 0 && c.Lines[len(c.Lines)-1].Line == line {
		c.Lines[len(c.Lines)-1].Count++
	} else {
		c.Lines = append(c.Lines, LineRun{Line: line, Count: 1})
	}
}
