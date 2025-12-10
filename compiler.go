package main

import (
	"fmt"

	"github.com/caelondev/hydor/tokens"
)

type Parser struct {}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Compile(source string) {
	scanner := NewScanner(source)

	line := -1
	for {
		token := scanner.ScanToken()

		if token.Line != line {
			fmt.Printf("%4d | ", token.Line)
			line = token.Line
		} else {
			fmt.Printf("   ^ ")
		}

		fmt.Printf("(%2d) '%s'\n", token.Type, token.Lexeme)

		if token.Type == tokens.TOKEN_EOF {
			break
		}
	}
}
