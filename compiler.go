package main

import (
	"fmt"
	"strconv"

	"github.com/caelondev/hydor/tokens"
)

const UINT8_MAX = 16 * 1024 * 1024
const DEBUG_PRINT_BYTECODE = false

type Parser struct {
	scanner     *Scanner
	current     tokens.Token
	previous    tokens.Token
	hadError    bool
	panicMode   bool
	compilingChunk *Chunk  // ← Added this
}

type Precedence int
type ParseFn func()

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

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Compile(source string, chunk *Chunk) bool {
	p.scanner = NewScanner(source)
	p.compilingChunk = chunk  // ← Store the chunk
	p.hadError = false
	p.panicMode = false
	
	p.advance()
	p.expression()
	p.consume(tokens.TOKEN_EOF, "Expected end of file")
	p.endCompiler()  // ← Add this
	
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

func (p *Parser) errorAtCurrent(message string) {
	p.errorAt(&p.current, message)
}

func (p *Parser) errorAt(token *tokens.Token, message string) {
	if p.panicMode {
		return
	}
	p.panicMode = true
	p.hadError = true

	fmt.Printf("[line %d] Error", token.Line)

	if token.Type == tokens.TOKEN_EOF {
		fmt.Printf(" at end")
	} else if token.Type == tokens.TOKEN_ERROR {
		// Nothing
	} else {
		fmt.Printf(" at '%s'", token.Lexeme)
	}

	fmt.Printf(": %s\n", message)
}

func (p *Parser) error(message string) {
	p.errorAt(&p.previous, message)
}

func (p *Parser) consume(tokenType tokens.TokenType, message string) {
	if p.current.Type == tokenType {
		p.advance()
		return
	}

	p.errorAtCurrent(message)
}

func (p *Parser) expression() {
	p.parsePrecedence(PREC_ASSIGNMENT)
}

func (p *Parser) emitByte(b byte) {
	p.compilingChunk.Write(b, p.previous.Line)
}

func (p *Parser) emitBytes(b1, b2 byte) {
	p.emitByte(b1)
	p.emitByte(b2)
}

func (p *Parser) emitReturn() {
	p.emitByte(byte(OP_RETURN))
	if !p.hadError && DEBUG_PRINT_BYTECODE {
		p.currentChunk().DisassembleChunk("Bytecode")
	}
}

func (p *Parser) emitConstant(value Value) {
	p.emitBytes(byte(OP_CONSTANT), p.makeConstant(value))
}

func (p *Parser) makeConstant(value Value) byte {
	constant := p.currentChunk().AddConstant(value)
	if constant > UINT8_MAX {
		p.error("Too many constants in one chunk.")
		return 0
	}

	return byte(constant)
}

func (p *Parser) endCompiler() {
	p.emitReturn()
}

func (p *Parser) number() {
	value, err := strconv.ParseFloat(p.previous.Lexeme, 64)
	if err != nil {
		// NOTE: This shouldn't happen since the scanner validated it
		// But just in case...
		p.error("Invalid number.")
		return
	}
	p.emitConstant(Value(value))
}

func (p *Parser) binary() {
	operatorType := p.previous.Type
	rule := p.getRule(operatorType)
	p.parsePrecedence(Precedence(rule.precedence))

	switch operatorType {
	case tokens.TOKEN_PLUS:
		p.emitByte(byte(OP_ADD))
	case tokens.TOKEN_MINUS:
		p.emitByte(byte(OP_SUBTRACT))
	case tokens.TOKEN_STAR:
		p.emitByte(byte(OP_MULTIPLY))
	case tokens.TOKEN_SLASH:
		p.emitByte(byte(OP_DIVIDE))
	default:
		return // Unreachable
	}
}

func (p *Parser) grouping() {
	p.expression()
	p.consume(tokens.TOKEN_RIGHT_PAREN, "Expected ')' after a grouping expression")
}

func (p *Parser) unary() {
	operatorType := p.previous.Type

	p.parsePrecedence(PREC_UNARY)

	switch operatorType {
	case tokens.TOKEN_MINUS: p.emitByte(byte(OP_NEGATE))
	default: return // Unreachable code
	}
}

func (p *Parser) getRule(tokenType tokens.TokenType) *ParseRule {
	rules := map[tokens.TokenType]ParseRule{
		tokens.TOKEN_LEFT_PAREN:    {p.grouping, nil, PREC_NONE},
		tokens.TOKEN_RIGHT_PAREN:   {nil, nil, PREC_NONE},
		tokens.TOKEN_LEFT_BRACE:    {nil, nil, PREC_NONE},
		tokens.TOKEN_RIGHT_BRACE:   {nil, nil, PREC_NONE},
		tokens.TOKEN_COMMA:         {nil, nil, PREC_NONE},
		tokens.TOKEN_DOT:           {nil, nil, PREC_NONE},
		tokens.TOKEN_MINUS:         {p.unary, p.binary, PREC_TERM},
		tokens.TOKEN_PLUS:          {nil, p.binary, PREC_TERM},
		tokens.TOKEN_SEMICOLON:     {nil, nil, PREC_NONE},
		tokens.TOKEN_SLASH:         {nil, p.binary, PREC_FACTOR},
		tokens.TOKEN_STAR:          {nil, p.binary, PREC_FACTOR},
		tokens.TOKEN_BANG:          {nil, nil, PREC_NONE},
		tokens.TOKEN_BANG_EQUAL:    {nil, nil, PREC_NONE},
		tokens.TOKEN_EQUAL:         {nil, nil, PREC_NONE},
		tokens.TOKEN_EQUAL_EQUAL:   {nil, nil, PREC_NONE},
		tokens.TOKEN_GREATER:       {nil, nil, PREC_NONE},
		tokens.TOKEN_GREATER_EQUAL: {nil, nil, PREC_NONE},
		tokens.TOKEN_LESS:          {nil, nil, PREC_NONE},
		tokens.TOKEN_LESS_EQUAL:    {nil, nil, PREC_NONE},
		tokens.TOKEN_IDENTIFIER:    {nil, nil, PREC_NONE},
		tokens.TOKEN_STRING:        {nil, nil, PREC_NONE},
		tokens.TOKEN_NUMBER:        {p.number, nil, PREC_NONE},
		tokens.TOKEN_AND:           {nil, nil, PREC_NONE},
		tokens.TOKEN_CLASS:         {nil, nil, PREC_NONE},
		tokens.TOKEN_ELSE:          {nil, nil, PREC_NONE},
		tokens.TOKEN_FALSE:         {nil, nil, PREC_NONE},
		tokens.TOKEN_FOR:           {nil, nil, PREC_NONE},
		tokens.TOKEN_FUNCTION:      {nil, nil, PREC_NONE},
		tokens.TOKEN_IF:            {nil, nil, PREC_NONE},
		tokens.TOKEN_NIL:           {nil, nil, PREC_NONE},
		tokens.TOKEN_OR:            {nil, nil, PREC_NONE},
		tokens.TOKEN_PRINT:         {nil, nil, PREC_NONE},
		tokens.TOKEN_RETURN:        {nil, nil, PREC_NONE},
		tokens.TOKEN_SUPER:         {nil, nil, PREC_NONE},
		tokens.TOKEN_THIS:          {nil, nil, PREC_NONE},
		tokens.TOKEN_TRUE:          {nil, nil, PREC_NONE},
		tokens.TOKEN_VAR:           {nil, nil, PREC_NONE},
		tokens.TOKEN_WHILE:         {nil, nil, PREC_NONE},
		tokens.TOKEN_ERROR:         {nil, nil, PREC_NONE},
		tokens.TOKEN_EOF:           {nil, nil, PREC_NONE},
	}

	rule, ok := rules[tokenType]
	if !ok {
		return &ParseRule{nil, nil, PREC_NONE}
	}
	return &rule
}

func (p *Parser) parsePrecedence(precedence Precedence) {
	p.advance()

	prefixRule := p.getRule(p.previous.Type).prefix
	if prefixRule == nil {
		p.error("Expected expression")
		return
	}

	prefixRule()

	for precedence <= p.getRule(p.current.Type).precedence {
		p.advance()
		infixRule := p.getRule(p.previous.Type).infix

		infixRule()
	}
}

func (p *Parser) currentChunk() *Chunk {
	return p.compilingChunk
}
