package bytecode

import "github.com/caelondev/hydor/runtime/value"

type OpCode byte
const (
	OP_CONSTANT OpCode = iota
	OP_NIL
	OP_TRUE
	OP_FALSE
	OP_EQUAL
	OP_GREATER
	OP_LESS
	OP_NOT
	OP_ADD
	OP_SUBTRACT
	OP_MULTIPLY
	OP_DIVIDE
	OP_MODULO
	OP_NEGATE
	OP_RETURN
)

type Bytecode struct {
	Code      []byte
	Lines     []LineRun
	Constants value.ValueArray
}

type LineRun struct { Line, Count int }

func NewBytecode(src string) *Bytecode {
	return &Bytecode{
		Code:      make([]byte, 0, len(src)+16),
		Lines:     make([]LineRun, 0, len(src)/4),
		Constants: *value.NewValueArray(),
	}
}

func (c *Bytecode) AddConstant(v value.Value) int {
	c.Constants.Write(v)
	return len(c.Constants.Values) - 1
}

func (c *Bytecode) Write(b byte, line int) {
	c.Code = append(c.Code, b)
	if len(c.Lines) > 0 && c.Lines[len(c.Lines)-1].Line == line {
		c.Lines[len(c.Lines)-1].Count++
	} else {
		c.Lines = append(c.Lines, LineRun{Line: line, Count: 1})
	}
}
