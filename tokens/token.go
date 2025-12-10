package tokens

type Token struct {
	Type   TokenType
	Start  int
	Length int
	Line   int
	Lexeme string
}
