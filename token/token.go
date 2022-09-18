package token

type TokenType string

const (
	ILLEGAL TokenType = "ILLEGAL"
	EOF               = "EOF"

	// Identifiers
	IDENT = "IDENT"

	// Literals
	INT = "INT"

	// Operators
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	LT       = "<"
	GT       = ">"

	// Delimiters
	COMMA     = ","
	SEMICOLON = ";"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"

	// Keywords
	FUNCTION = "FUNCTION"
	LET      = "LET"
)

var keywords = map[string]TokenType{
	"fn":  FUNCTION,
	"let": LET,
}

type Token struct {
	Type    TokenType
	Literal string
}

func GetIdentifierType(ident string) TokenType {
	if t, ok := keywords[ident]; ok {
		return t
	}
	return IDENT
}
