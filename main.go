package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/caelondev/hydor/frontend/bytecode"
	"github.com/caelondev/hydor/frontend/lexer"
	"github.com/caelondev/hydor/frontend/parser"
	"github.com/caelondev/hydor/result"
	"github.com/caelondev/hydor/runtime/vm"
)

func main() {
	if len(os.Args) == 1 {
		runRepl()
	} else if len(os.Args) == 2 {
		runFile(os.Args[1])
	} else {
		fmt.Fprintln(os.Stderr, "Usage: hydor [script]")
		os.Exit(64)
	}
}

func runRepl() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Hydor REPL - Type '/exit' to quit")

	for {
		fmt.Print(">> ")

		if !scanner.Scan() {
			fmt.Println()
			break
		}

		line := scanner.Text()

		if line == "/exit" {
			break
		}

		run(line)
	}
}

func runFile(path string) {
	source, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot open file '%s', Error: %s", path,  err.Error())
	}

	returnValue := run(string(source))

	if returnValue == result.INTERPRET_COMPILE_ERROR {
		os.Exit(65)
	}

	if returnValue == result.INTERPRET_RUNTIME_ERROR {
		os.Exit(64)
	}
}

func run(source string) result.InterpretResult {
	tokenizer := lexer.NewTokenizer(source)
	bytecode := bytecode.NewBytecode(source)
	parser := parser.NewParser()
	vm := vm.NewVM()
	
	if !parser.Compile(source, tokenizer, bytecode) {
		return result.INTERPRET_COMPILE_ERROR
	}

	result := vm.Interpret(bytecode)
	return result
}
