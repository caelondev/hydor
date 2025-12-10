package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Now()
	chunk := NewChunk()

	con := chunk.AddConstant(67)

	chunk.Write(byte(OP_CONSTANT), 1)
	chunk.Write(byte(con), 1)

	chunk.Write(byte(OP_RETURN), 1)

	vm := NewVM()
	vm.Interpret(chunk)

	fmt.Printf("Time was %s\n", time.Since(start))
}
