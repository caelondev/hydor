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
	PREC_parseUnary
	PREC_CALL
	PREC_PRIMARY
)

var parseRules map[tokens.TokenType]ParseRule

func init() {
	parseRules = map[tokens.TokenType]ParseRule{
		tokens.TOKEN_LEFT_PAREN:    {parseGrouping, nil, PREC_NONE},
		tokens.TOKEN_RIGHT_PAREN:   {nil, nil, PREC_NONE},
		tokens.TOKEN_LEFT_BRACE:    {nil, nil, PREC_NONE},
		tokens.TOKEN_RIGHT_BRACE:   {nil, nil, PREC_NONE},
		tokens.TOKEN_COMMA:         {nil, nil, PREC_NONE},
		tokens.TOKEN_DOT:           {nil, nil, PREC_NONE},
		tokens.TOKEN_MINUS:         {parseUnary, parseBinary, PREC_TERM},
		tokens.TOKEN_PLUS:          {nil, parseBinary, PREC_TERM},
		tokens.TOKEN_SEMICOLON:     {nil, nil, PREC_NONE},
		tokens.TOKEN_SLASH:         {nil, parseBinary, PREC_FACTOR},
		tokens.TOKEN_STAR:          {nil, parseBinary, PREC_FACTOR},
		tokens.TOKEN_BANG:          {parseUnary, nil, PREC_NONE},
		tokens.TOKEN_BANG_EQUAL:    {nil, parseBinary, PREC_EQUALITY},
		tokens.TOKEN_EQUAL:         {nil, nil, PREC_NONE},
		tokens.TOKEN_EQUAL_EQUAL:   {nil, parseBinary, PREC_EQUALITY},
		tokens.TOKEN_GREATER:       {nil, parseBinary, PREC_COMPARISON},
		tokens.TOKEN_GREATER_EQUAL: {nil, parseBinary, PREC_COMPARISON},
		tokens.TOKEN_LESS:          {nil, parseBinary, PREC_COMPARISON},
		tokens.TOKEN_LESS_EQUAL:    {nil, parseBinary, PREC_COMPARISON},
		tokens.TOKEN_IDENTIFIER:    {nil, nil, PREC_NONE},
		tokens.TOKEN_STRING:        {parseString, nil, PREC_NONE},
		tokens.TOKEN_NUMBER:        {parseNumber, nil, PREC_NONE},
		tokens.TOKEN_AND:           {nil, nil, PREC_NONE},
		tokens.TOKEN_CLASS:         {nil, nil, PREC_NONE},
		tokens.TOKEN_ELSE:          {nil, nil, PREC_NONE},
		tokens.TOKEN_FALSE:         {parseLiteral, nil, PREC_NONE},
		tokens.TOKEN_FOR:           {nil, nil, PREC_NONE},
		tokens.TOKEN_FUNCTION:      {nil, nil, PREC_NONE},
		tokens.TOKEN_IF:            {nil, nil, PREC_NONE},
		tokens.TOKEN_NIL:           {parseLiteral, nil, PREC_NONE},
		tokens.TOKEN_OR:            {nil, nil, PREC_NONE},
		tokens.TOKEN_PRINT:         {nil, nil, PREC_NONE},
		tokens.TOKEN_RETURN:        {nil, nil, PREC_NONE},
		tokens.TOKEN_SUPER:         {nil, nil, PREC_NONE},
		tokens.TOKEN_THIS:          {nil, nil, PREC_NONE},
		tokens.TOKEN_TRUE:          {parseLiteral, nil, PREC_NONE},
		tokens.TOKEN_VAR:           {nil, nil, PREC_NONE},
		tokens.TOKEN_WHILE:         {nil, nil, PREC_NONE},
		tokens.TOKEN_ERROR:         {nil, nil, PREC_NONE},
		tokens.TOKEN_EOF:           {nil, nil, PREC_NONE},
	}
}

type Parser struct {
	scanner             *Scanner
	current, previous   tokens.Token
	hadError, panicMode bool
	compilingChunk      *Chunk
}

func NewParser() *Parser {
	return &Parser{}
}

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

func parseString(p *Parser) {
	str := NewString(p.previous.Lexeme)
	p.emitConstant(ObjVal(str.AsObj()))
}

func parseNumber(p *Parser) {
	val, _ := strconv.ParseFloat(p.previous.Lexeme, 64)
	p.emitConstant(NumberVal(val))
}

func parseUnary(p *Parser) {
	op := p.previous.Type
	p.parsePrecedence(PREC_parseUnary)

	switch op {
	case tokens.TOKEN_MINUS: p.emitByte(byte(OP_NEGATE))
	case tokens.TOKEN_BANG: p.emitByte(byte(OP_NOT))
	}
}

func parseBinary(p *Parser) {
	op := p.previous.Type
	rule := parseRules[op]
	p.parsePrecedence(rule.precedence +1)
	switch op {
	case tokens.TOKEN_PLUS:
		p.emitByte(byte(OP_ADD))
	case tokens.TOKEN_MINUS:
		p.emitByte(byte(OP_SUBTRACT))
	case tokens.TOKEN_STAR:
		p.emitByte(byte(OP_MULTIPLY))
	case tokens.TOKEN_SLASH:
		p.emitByte(byte(OP_DIVIDE))

	// !(a == b)
	case tokens.TOKEN_BANG_EQUAL: p.emitBytes(byte(OP_EQUAL), byte(OP_NOT))
	case tokens.TOKEN_EQUAL_EQUAL: p.emitByte(byte(OP_EQUAL))
	case tokens.TOKEN_GREATER: p.emitByte(byte(OP_GREATER))
	// !(a < b)
	case tokens.TOKEN_GREATER_EQUAL: p.emitBytes(byte(OP_LESS), byte(OP_NOT))
	case tokens.TOKEN_LESS: p.emitByte(byte(OP_LESS))
	// !(a > b)
	case tokens.TOKEN_LESS_EQUAL: p.emitBytes(byte(OP_GREATER), byte(OP_NOT))
	}
}

func parseLiteral(p *Parser) {
	switch p.previous.Type {
	case tokens.TOKEN_FALSE:
		p.emitByte(byte(OP_FALSE))
	case tokens.TOKEN_TRUE:
		p.emitByte(byte(OP_TRUE))
	case tokens.TOKEN_NIL:
		p.emitByte(byte(OP_NIL))
	}
}

func parseGrouping(p *Parser) {
	p.expression()
	p.consume(tokens.TOKEN_RIGHT_PAREN, "Expected ')' after parseGrouping")
}
