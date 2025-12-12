package lexer

import (
	"fmt"

	"github.com/caelondev/hydor/frontend/tokens"
)

type Tokenizer struct {
	Source  string
	Start   int
	Current int
	Line    int
}

func NewTokenizer(source string) *Tokenizer {
	return &Tokenizer{
		Source:  source,
		Start:   0,
		Current: 0,
		Line:    1,
	}
}

func (s *Tokenizer) ScanToken() tokens.Token {
	s.skipIgnored()
	s.Start = s.Current

	if s.isAtEnd() {
		return s.newToken(tokens.TOKEN_EOF)
	}

	c := s.advance()

	if isAlphabet(c) {
		return s.identifier()
	}

	if isDigit(c) {
		return s.number()
	}

	switch c {
	case '(': return s.newToken(tokens.TOKEN_LEFT_PAREN)
	case ')': return s.newToken(tokens.TOKEN_RIGHT_PAREN)
	case '{': return s.newToken(tokens.TOKEN_LEFT_BRACE)
	case '}': return s.newToken(tokens.TOKEN_RIGHT_BRACE)
	case ';': return s.newToken(tokens.TOKEN_SEMICOLON)
	case ',': return s.newToken(tokens.TOKEN_COMMA)
	case '.': return s.newToken(tokens.TOKEN_DOT)
	case '-': return s.newToken(tokens.TOKEN_MINUS)
	case '+': return s.newToken(tokens.TOKEN_PLUS)
	case '/': return s.newToken(tokens.TOKEN_SLASH)
	case '*': return s.newToken(tokens.TOKEN_STAR)
	case '!': return s.matchEqual(tokens.TOKEN_BANG, tokens.TOKEN_BANG_EQUAL)
	case '<': return s.matchEqual(tokens.TOKEN_LESS, tokens.TOKEN_LESS_EQUAL)
	case '>': return s.matchEqual(tokens.TOKEN_GREATER, tokens.TOKEN_GREATER_EQUAL)
	case '=': return s.matchEqual(tokens.TOKEN_EQUAL, tokens.TOKEN_EQUAL_EQUAL)
	case '"', '\'':
		return s.string(c)
	case '`':
		return s.multilineString(c)
	}

	return s.errorToken(fmt.Sprintf("Unknown character found '%c'", c))
}

func (s *Tokenizer) multilineString(terminator byte) tokens.Token {
	startLine := s.Line
	for s.peek() != terminator && !s.isAtEnd() {
		if s.peek() == '\n' { s.Line++ }
		s.advance()
	}

	if s.isAtEnd() {
		return s.errorToken("Unterminated multi-line string")
	}
	s.advance()
	lexeme := s.Source[s.Start+1 : s.Current-1]
	return tokens.Token{
		Type: tokens.TOKEN_STRING,
		Start: s.Start,
		Line: startLine,
		Length: s.Line,
		Lexeme: lexeme,
	}
}

func (s *Tokenizer) identifier() tokens.Token {
	for isAlphanumeric(s.peek()) { s.advance() }
	lexeme := s.Source[s.Start:s.Current]
	if kw, ok := tokens.RESERVED_KEYWORDS[lexeme]; ok {
		return s.newToken(kw)
	}

	return s.newToken(tokens.TOKEN_IDENTIFIER)
}

func (s *Tokenizer) number() tokens.Token {
	for isDigit(s.peek()) { s.advance() }
	if s.peek() == '.' && isDigit(s.peekNext()) {
		s.advance()

		for isDigit(s.peek()) {
			s.advance()
		}
	}
	return s.newToken(tokens.TOKEN_NUMBER)
}

func (s *Tokenizer) string(terminator byte) tokens.Token {
	for s.peek() != terminator && s.peek() != '\n' && !s.isAtEnd() { s.advance() }
	if s.isAtEnd() || s.peek() == '\n' {
		return s.errorToken("Unterminated non-multiline string")
	}

	s.advance()
	lexeme := s.Source[s.Start+1 : s.Current-1]
	return s.newTokenLexeme(tokens.TOKEN_STRING, lexeme)
}

func (s *Tokenizer) skipIgnored() {
	for {
		c := s.peek()
		switch c {
		case ' ', '\r', '\t':
			s.advance()
		case '\n':
			s.Line++
			s.advance()
		case '/':
			if s.peekNext() == '/' {
				s.advance(); s.advance()
				for s.peek() != '\n' && !s.isAtEnd() { s.advance() }
			} else if s.peekNext() == '*' {
				s.advance(); s.advance()
				for !s.isAtEnd() && !(s.peek() == '*' && s.peekNext() == '/') {
					if s.peek() == '\n' { s.Line++ }
					s.advance()
				}
				if s.isAtEnd() { return }
				s.advance(); s.advance()
			} else { return }
		default:
			return
		}
	}
}

func (s *Tokenizer) peek() byte {
	if s.isAtEnd() {
		return 0
	}

	return s.Source[s.Current]
}

func (s *Tokenizer) peekNext() byte {
	if s.Current+1 >= len(s.Source) {
		return 0
	}

	return s.Source[s.Current+1]
}

func (s *Tokenizer) advance() byte {
	s.Current++
	return s.Source[s.Current-1]
}

func (s *Tokenizer) match(expected byte) bool {
	if s.isAtEnd() || s.Source[s.Current] != expected {
		return false
	}

	s.Current++
	return true
}

func (s *Tokenizer) matchEqual(single, eq tokens.TokenType) tokens.Token {
	if s.match('=') {
		return s.newToken(eq)
	}

	return s.newToken(single)
}

func (s *Tokenizer) newToken(tt tokens.TokenType) tokens.Token {
	return s.newTokenLexeme(tt, s.Source[s.Start:s.Current])
}

func (s *Tokenizer) newTokenLexeme(tt tokens.TokenType, lexeme string) tokens.Token {
	return tokens.Token{
		Type: tt,
		Line: s.Line,
		Start: s.Start,
		Length: len(lexeme),
		Lexeme: lexeme,
	}
}

func (s *Tokenizer) errorToken(msg string) tokens.Token {
	return s.newTokenLexeme(tokens.TOKEN_ERROR, msg)
}

func (s *Tokenizer) isAtEnd() bool {
	return s.Current >= len(s.Source)
}

func isAlphanumeric(c byte) bool {
	return isAlphabet(c) || isDigit(c)
}
func isAlphabet(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
	       (c >= 'A' && c <= 'Z') || 
	        isUnderscore(c)
}

func isUnderscore(c byte) bool {
	return c == '_'
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }
