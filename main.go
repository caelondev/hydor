package main

import (
	"bufio"
	"fmt"
	"os"
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
	vm := NewVM()

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

		if line == "" {
			continue
		}

		vm.Interpret(line)
	}
}

func runFile(path string) {
	source, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot open file '%s', Error: %s", path,  err.Error())
	}

	vm := NewVM()
	result := vm.Interpret(string(source))

	if result == INTERPRET_COMPILE_ERROR {
		os.Exit(65)
	}

	if result == INTERPRET_RUNTIME_ERROR {
		os.Exit(64)
	}
}
