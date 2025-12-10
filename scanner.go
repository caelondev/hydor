package main

import (
	"fmt"

	"github.com/caelondev/hydor/tokens"
)

type Scanner struct {
	Source  string
	Start   int
	Current int
	Line    int
}

func NewScanner(source string) *Scanner {
	return &Scanner{
		Source:  source,
		Start:   0,
		Current: 0,
		Line:    1,
	}
}

func (s *Scanner) ScanToken() tokens.Token {
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
	case '(':
		return s.newToken(tokens.TOKEN_LEFT_PAREN)
	case ')':
		return s.newToken(tokens.TOKEN_RIGHT_PAREN)
	case '{':
		return s.newToken(tokens.TOKEN_LEFT_BRACE)
	case '}':
		return s.newToken(tokens.TOKEN_RIGHT_BRACE)
	case ';':
		return s.newToken(tokens.TOKEN_SEMICOLON)
	case ',':
		return s.newToken(tokens.TOKEN_COMMA)
	case '.':
		return s.newToken(tokens.TOKEN_DOT)
	case '-':
		return s.newToken(tokens.TOKEN_MINUS)
	case '+':
		return s.newToken(tokens.TOKEN_PLUS)
	case '/':
		return s.newToken(tokens.TOKEN_SLASH)
	case '*':
		return s.newToken(tokens.TOKEN_STAR)
	case '!':
		return s.matchEqual(tokens.TOKEN_BANG, tokens.TOKEN_BANG_EQUAL)
	case '<':
		return s.matchEqual(tokens.TOKEN_LESS, tokens.TOKEN_LESS_EQUAL)
	case '>':
		return s.matchEqual(tokens.TOKEN_GREATER, tokens.TOKEN_GREATER_EQUAL)
	case '=':
		return s.matchEqual(tokens.TOKEN_EQUAL, tokens.TOKEN_EQUAL_EQUAL)
	case '"', '\'':
		return s.string(c)
	case '`':
		return s.multilineString(c)
	}

	return s.errorToken(fmt.Sprintf("Unknown character found '%c'", c))
}

func (s *Scanner) multilineString(terminator byte) tokens.Token {
	startLine := s.Line // Cache actual string line

	for s.peek() != terminator && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.Line++
		}

		s.advance()
	}

	if s.isAtEnd() {
		return s.errorToken("Unterminated multi-line string")
	}

	s.advance() // Eat terminator

	lexeme := s.Source[s.Start +1 : s.Current -1]
	return tokens.Token{
		Type:   tokens.TOKEN_STRING,
		Line:   startLine,
		Start:  s.Start,
		Length: len(lexeme),
		Lexeme: lexeme,
	}
}

func (s *Scanner) identifier() tokens.Token {
	for isAlphanumeric(s.peek()) {
		s.advance()
	}

	lexeme := s.Source[s.Start:s.Current]
	keyword, isReserved := tokens.RESERVED_KEYWORDS[lexeme]

	if isReserved {
		return s.newToken(keyword)
	}
	return s.newToken(tokens.TOKEN_IDENTIFIER)
}

func (s *Scanner) number() tokens.Token {
	for isDigit(s.peek()) {
		s.advance()
	}

	if s.peek() == '.' && isDigit(s.peekNext()) {
		s.advance() // Consume '.'

		for isDigit(s.peek()) {
			s.advance()
		}
	}

	return s.newToken(tokens.TOKEN_NUMBER)
}

func (s *Scanner) string(terminator byte) tokens.Token {
	for s.peek() != terminator && s.peek() != '\n' && !s.isAtEnd() {
		s.advance()
	}

	if s.isAtEnd() || s.peek() == '\n' {
		return s.errorToken("Unterminated non-multiline string")
	}

	s.advance() // Consume closing quote
	lexeme := s.Source[s.Start +1 : s.Current -1] // Exclude quote tokens
	return s.newTokenLexeme(tokens.TOKEN_STRING, lexeme)
}

func (s *Scanner) isAtEnd() bool {
	return s.Current >= len(s.Source)
}

func isAlphanumeric(char byte) bool {
	return isAlphabet(char) || isDigit(char)
}

func isAlphabet(char byte) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		char == '_'
}

func isDigit(char byte) bool {
	return char >= '0' && char <= '9'
}

func (s *Scanner) skipIgnored() {
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
				// Consume '//'
				s.advance()
				s.advance()
				for s.peek() != '\n' && !s.isAtEnd() {
					s.advance()
				}
			} else if s.peekNext() == '*' {
				// Consume '/*'
				s.advance()
				s.advance()

				for !s.isAtEnd() && !(s.peek() == '*' && s.peekNext() == '/') {
					if s.peek() == '\n' {
						s.Line++
					}
					s.advance()
				}

				if s.isAtEnd() {
					// Unterminated block comment
					return
				}

				// Consume '*/'
				s.advance()
				s.advance()
			} else {
				return
			}

		default:
			return
		}
	}
}

func (s *Scanner) peek() byte {
	if s.isAtEnd() {
		return 0
	}
	return s.Source[s.Current]
}

func (s *Scanner) peekNext() byte {
	if s.Current+1 >= len(s.Source) { // ← FIXED: Check Current+1
		return 0
	}
	return s.Source[s.Current+1]
}

func (s *Scanner) match(expected byte) bool {
	if s.isAtEnd() || s.Source[s.Current] != expected {
		return false
	}
	s.Current++
	return true
}

func (s *Scanner) matchEqual(single tokens.TokenType, withEqual tokens.TokenType) tokens.Token {
	if s.match('=') {
		return s.newToken(withEqual)
	}
	return s.newToken(single)
}

func (s *Scanner) newToken(tokenType tokens.TokenType) tokens.Token {
	return s.newTokenLexeme(tokenType, s.Source[s.Start:s.Current])
}

func (s *Scanner) newTokenLexeme(tokenType tokens.TokenType, lexeme string) tokens.Token {
	return tokens.Token{
		Type:   tokenType,
		Line:   s.Line,
		Start:  s.Start,
		Length: len(lexeme),
		Lexeme: lexeme,
	}
}

func (s *Scanner) errorToken(message string) tokens.Token {
	return s.newTokenLexeme(tokens.TOKEN_ERROR, message)
}

func (s *Scanner) advance() byte {
	s.Current++
	return s.Source[s.Current-1]
}
