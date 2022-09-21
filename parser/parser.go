package parser

import (
	"fmt"
	"strconv"

	"protiumx.dev/simia/ast"
	"protiumx.dev/simia/lexer"
	"protiumx.dev/simia/token"
)

// Precedences
const (
	_ int = iota
	LOWEST
	EQUALS
	LESS_GREATER
	SUM
	PRODUCT
	PREFIX
	CALL
)

type Parser struct {
	lexer          *lexer.Lexer
	currentToken   token.Token
	peekToken      token.Token
	errors         []string
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

func New(l *lexer.Lexer) *Parser {
	p := Parser{lexer: l, errors: []string{}}
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)

	// Set values for current and peak
	p.nextToken()
	p.nextToken()
	return &p
}

func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) ParseProgram() *ast.Program {
	program := ast.NewProgram()

	for !p.currentTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) registerPrefix(t token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[t] = fn
}

func (p *Parser) registerInfix(t token.TokenType, fn infixParseFn) {
	p.infixParseFns[t] = fn
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.currentToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.currentToken}
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	for !p.currentTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatment {
	stmt := &ast.ReturnStatment{Token: p.currentToken}
	p.nextToken()

	for !p.currentTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatment {
	stmt := &ast.ExpressionStatment{Token: p.currentToken}
	stmt.Expression = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.currentToken.Type]
	if prefix == nil {
		return nil
	}

	return prefix()
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	literal := &ast.IntigerLiteral{Token: p.currentToken}
	value, err := strconv.ParseInt(p.currentToken.Literal, 0, 64)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("could not parse %q as integer", p.currentToken))
		return nil
	}
	literal.Value = value
	return literal
}

func (p *Parser) currentTokenIs(t token.TokenType) bool {
	return p.currentToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}

	p.peekError(t)
	return false
}

func (p *Parser) peekError(expectedToken token.TokenType) {
	p.errors = append(p.errors, fmt.Sprintf("expected next token to be %s, got %s", expectedToken, p.peekToken.Type))
}
