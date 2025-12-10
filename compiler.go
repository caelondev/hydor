package main

import (
	"fmt"
	"strconv"

	"github.com/caelondev/hydor/tokens"
)

const UINT8_MAX = 255

type Precedence int
type ParseFn func(*Parser)

type ParseRule struct {
	prefix     ParseFn
	infix      ParseFn
	precedence Precedence
}

const (
	PREC_NONE Precedence = iota
	PREC_ASSIGNMENT
	PREC_OR
	PREC_AND
	PREC_EQUALITY
	PREC_COMPARISON
	PREC_TERM
	PREC_FACTOR
	PREC_UNARY
	PREC_CALL
	PREC_PRIMARY
)

var parseRules map[tokens.TokenType]ParseRule

func init() {
	parseRules = map[tokens.TokenType]ParseRule{
		tokens.TOKEN_LEFT_PAREN:  {grouping, nil, PREC_NONE},
		tokens.TOKEN_RIGHT_PAREN: {nil, nil, PREC_NONE},
		tokens.TOKEN_PLUS:        {nil, binary, PREC_TERM},
		tokens.TOKEN_MINUS:       {unary, binary, PREC_TERM},
		tokens.TOKEN_STAR:        {nil, binary, PREC_FACTOR},
		tokens.TOKEN_SLASH:       {nil, binary, PREC_FACTOR},
		tokens.TOKEN_NUMBER:      {number, nil, PREC_NONE},
	}
}

type Parser struct {
	scanner             *Scanner
	current, previous   tokens.Token
	hadError, panicMode bool
	compilingChunk      *Chunk
}

func NewParser() *Parser { return &Parser{} }

func (p *Parser) Compile(source string, chunk *Chunk) bool {
	p.scanner = NewScanner(source)
	p.compilingChunk = chunk
	p.hadError = false
	p.panicMode = false

	p.advance()
	p.expression()
	p.consume(tokens.TOKEN_EOF, "Expected end of file")
	p.endCompiler()

	return !p.hadError
}

func (p *Parser) advance() {
	p.previous = p.current
	for {
		p.current = p.scanner.ScanToken()
		if p.current.Type != tokens.TOKEN_ERROR {
			break
		}
		p.errorAtCurrent(p.current.Lexeme)
	}
}

func (p *Parser) consume(tt tokens.TokenType, msg string) {
	if p.current.Type == tt {
		p.advance()
		return
	}
	p.errorAtCurrent(msg)
}

func (p *Parser) error(msg string) {
	p.errorAt(&p.previous, msg)
}

func (p *Parser) errorAtCurrent(msg string) {
	p.errorAt(&p.current, msg)
}

func (p *Parser) errorAt(t *tokens.Token, msg string) {
	if p.panicMode {
		return
	}
	p.panicMode = true
	p.hadError = true

	fmt.Printf("[line %d] Error", t.Line)
	if t.Type == tokens.TOKEN_EOF {
		fmt.Printf(" at end")
	} else if t.Type != tokens.TOKEN_ERROR {
		fmt.Printf(" at '%s'", t.Lexeme)
	}
	fmt.Printf(": %s\n", msg)
}

func (p *Parser) expression() {
	p.parsePrecedence(PREC_ASSIGNMENT)
}

func (p *Parser) parsePrecedence(precedence Precedence) {
	p.advance()
	prefix := parseRules[p.previous.Type].prefix
	if prefix == nil {
		p.error("Expected expression")
		return
	}
	prefix(p)

	for precedence <= parseRules[p.current.Type].precedence {
		p.advance()
		infix := parseRules[p.previous.Type].infix
		infix(p)
	}
}

func (p *Parser) emitByte(b byte)       {
	p.compilingChunk.Write(b, p.previous.Line)
}

func (p *Parser) emitBytes(b1, b2 byte) { 
	p.emitByte(b1); p.emitByte(b2) 
}

func (p *Parser) emitReturn() { 
	p.emitByte(byte(OP_RETURN)) 
}

func (p *Parser) emitConstant(v Value) { 
	p.emitBytes(byte(OP_CONSTANT), p.makeConstant(v))
}

func (p *Parser) makeConstant(v Value) byte {
	idx := p.compilingChunk.AddConstant(v)
	if idx > UINT8_MAX {
		p.error("Too many constants")
		return 0
	}
	return byte(idx)
}

func (p *Parser) endCompiler() {
	p.emitReturn()
}

// ---- Parse functions ----

func number(p *Parser) {
	val, _ := strconv.ParseFloat(p.previous.Lexeme, 64)
	p.emitConstant(Value(val))
}

func unary(p *Parser) {
	op := p.previous.Type
	p.parsePrecedence(PREC_UNARY)
	if op == tokens.TOKEN_MINUS {
		p.emitByte(byte(OP_NEGATE))
	}
}

func binary(p *Parser) {
	op := p.previous.Type
	rule := parseRules[op]
	p.parsePrecedence(rule.precedence)
	switch op {
	case tokens.TOKEN_PLUS:
		p.emitByte(byte(OP_ADD))
	case tokens.TOKEN_MINUS:
		p.emitByte(byte(OP_SUBTRACT))
	case tokens.TOKEN_STAR:
		p.emitByte(byte(OP_MULTIPLY))
	case tokens.TOKEN_SLASH:
		p.emitByte(byte(OP_DIVIDE))
	}
}

func grouping(p *Parser) {
	p.expression()
	p.consume(tokens.TOKEN_RIGHT_PAREN, "Expected ')' after grouping")
}
