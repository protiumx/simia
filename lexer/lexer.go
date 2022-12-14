package lexer

import (
	"fmt"

	"protiumx.dev/simia/token"
)

type Lexer struct {
	input          string
	currentPostion int
	// TODO: support Unicode with `rune`
	// Value of 0 represents the NULL char
	currentChar  byte
	readPotition int
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) NextToken() token.Token {
	l.consumeWhiteSpace()
	var ret token.Token
	switch l.currentChar {
	case '=':
		ret = l.getOneOrTwoCharToken('=', token.ASSIGN, token.EQ)
	case '!':
		ret = l.getOneOrTwoCharToken('=', token.BANG, token.NOT_EQ)
	case '+':
		ret = newToken(token.PLUS, l.currentChar)
	case '-':
		ret = newToken(token.MINUS, l.currentChar)
	case '/':
		ret = newToken(token.SLASH, l.currentChar)
	case '*':
		ret = newToken(token.ASTERISK, l.currentChar)
	case '<':
		ret = newToken(token.LT, l.currentChar)
	case '>':
		ret = newToken(token.GT, l.currentChar)
	case ';':
		ret = newToken(token.SEMICOLON, l.currentChar)
	case ',':
		ret = newToken(token.COMMA, l.currentChar)
	case '(':
		ret = newToken(token.LPAREN, l.currentChar)
	case ')':
		ret = newToken(token.RPAREN, l.currentChar)
	case '{':
		ret = newToken(token.LBRACE, l.currentChar)
	case '}':
		ret = newToken(token.RBRACE, l.currentChar)
	case '[':
		ret = newToken(token.LBRACKET, l.currentChar)
	case ']':
		ret = newToken(token.RBRACKET, l.currentChar)
	case ':':
		ret = newToken(token.COLON, l.currentChar)
	case '|':
		ret = l.getTwoCharToken('>', token.PIPELINE)
	case '.':
		ret = l.getTwoCharToken('.', token.RANGE)
	case '"':
		ret.Type = token.STRING
		ret.Literal = l.readString()
	case 0:
		ret.Literal = ""
		ret.Type = token.EOF
	default:
		if isLetter(l.currentChar) {
			ret.Literal = l.readIdentifier()
			ret.Type = token.GetIdentifierType(ret.Literal)
			return ret
		} else if isDigit(l.currentChar) {
			ret.Literal = l.readNumber()
			ret.Type = token.INT
			return ret
		} else {
			ret = newToken(token.ILLEGAL, l.currentChar)
		}
	}
	l.readChar()
	return ret
}

func newToken(tokenType token.TokenType, char byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(char)}
}

func (l *Lexer) readChar() {
	if l.readPotition >= len(l.input) {
		l.currentChar = 0
	} else {
		l.currentChar = l.input[l.readPotition]
	}
	l.currentPostion = l.readPotition
	l.readPotition++
}

func (l *Lexer) readIdentifier() string {
	position := l.currentPostion
	for isLetter(l.currentChar) {
		l.readChar()
	}
	return l.input[position:l.currentPostion]
}

func (l *Lexer) readNumber() string {
	position := l.currentPostion
	for isDigit(l.currentChar) {
		l.readChar()
	}
	return l.input[position:l.currentPostion]
}

func (l *Lexer) readString() string {
	position := l.currentPostion + 1
	for {
		l.readChar()
		// TODO: check next token is valid (\t, \n, ...)
		if l.currentChar == '\\' {
			l.readChar()
			continue
		}

		if l.currentChar == '"' || l.currentChar == 0 {
			break
		}
	}
	return l.input[position:l.currentPostion]
}

func (l *Lexer) consumeWhiteSpace() {
	for l.currentChar == ' ' || l.currentChar == '\t' || l.currentChar == '\n' || l.currentChar == '\r' {
		l.readChar()
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPotition >= len(l.input) {
		return 0
	}
	return l.input[l.readPotition]
}

func (l *Lexer) getOneOrTwoCharToken(secondChar byte, oneCharType, twoCharType token.TokenType) token.Token {
	if l.peekChar() == secondChar {
		char := l.currentChar
		l.readChar()
		return token.Token{Type: twoCharType, Literal: fmt.Sprintf("%c%c", char, l.currentChar)}
	}
	return newToken(oneCharType, l.currentChar)
}

func (l *Lexer) getTwoCharToken(secondChar byte, tokenType token.TokenType) token.Token {
	if l.peekChar() != secondChar {
		return newToken(token.ILLEGAL, l.currentChar)
	}

	char := l.currentChar
	l.readChar()
	return token.Token{Type: tokenType, Literal: fmt.Sprintf("%c%c", char, l.currentChar)}
}

func isLetter(char byte) bool {
	return 'a' <= char && char <= 'z' || 'A' <= char && char <= 'Z' || char == '_'
}

func isDigit(char byte) bool {
	return '0' <= char && char <= '9'
}
