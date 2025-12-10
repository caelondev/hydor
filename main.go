package main

import (
	"fmt"
	"time"
)

func main() {
	chunk := NewChunk()
	
	for i := 0; i < 200_000; i++ {
		// a + b
		a := chunk.AddConstant(Value(i + 1))
		chunk.Write(byte(OP_CONSTANT), 1)
		chunk.Write(byte(a), 1)
		
		b := chunk.AddConstant(Value(i + 2))
		chunk.Write(byte(OP_CONSTANT), 1)
		chunk.Write(byte(b), 1)
		
		chunk.Write(byte(OP_ADD), 1)
		
		// c - d
		c := chunk.AddConstant(Value(i + 3))
		chunk.Write(byte(OP_CONSTANT), 1)
		chunk.Write(byte(c), 1)
		
		d := chunk.AddConstant(Value(i + 4))
		chunk.Write(byte(OP_CONSTANT), 1)
		chunk.Write(byte(d), 1)
		
		chunk.Write(byte(OP_SUBTRACT), 1)
		
		// Multiply results
		chunk.Write(byte(OP_MULTIPLY), 1)
		
		// e + f
		e := chunk.AddConstant(Value(i + 5))
		chunk.Write(byte(OP_CONSTANT), 1)
		chunk.Write(byte(e), 1)
		
		f := chunk.AddConstant(Value(i + 6))
		chunk.Write(byte(OP_CONSTANT), 1)
		chunk.Write(byte(f), 1)
		
		chunk.Write(byte(OP_ADD), 1)
		
		// Divide
		chunk.Write(byte(OP_DIVIDE), 1)
		
		// g * h
		g := chunk.AddConstant(Value(i + 7))
		chunk.Write(byte(OP_CONSTANT), 1)
		chunk.Write(byte(g), 1)
		
		h := chunk.AddConstant(Value(i + 8))
		chunk.Write(byte(OP_CONSTANT), 1)
		chunk.Write(byte(h), 1)
		
		chunk.Write(byte(OP_MULTIPLY), 1)
		
		// Final subtract
		chunk.Write(byte(OP_SUBTRACT), 1)
	}
	
	chunk.Write(byte(OP_RETURN), 1)

	vm := NewVM()
	
	start := time.Now()
	vm.Interpret(chunk)
	elapsed := time.Since(start)
	
	instructionCount := len(chunk.Code)
	nanosPerInstruction := float64(elapsed.Nanoseconds()) / float64(instructionCount)
	
	fmt.Printf("VM time: %s\n", elapsed)
	fmt.Printf("Instructions executed: %d\n", instructionCount)
	fmt.Printf("Time per instruction: %.2f ns/instruction\n", nanosPerInstruction)
	fmt.Printf("Instructions per second: %.2f million/sec\n", 1000.0/nanosPerInstruction)
}
