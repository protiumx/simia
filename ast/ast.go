package ast

import (
	"strings"

	"protiumx.dev/simia/token"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode() {}

func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

func (i *Identifier) String() string {
	return i.Value
}

type LetStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode() {}

func (ls *LetStatement) TokenLiteral() string {
	return ls.Token.Literal
}

func (ls *LetStatement) String() string {
	var out strings.Builder
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

type ReturnStatment struct {
	Token       token.Token
	ReturnValue Expression
}

func (rs *ReturnStatment) statementNode() {}

func (rs *ReturnStatment) TokenLiteral() string {
	return rs.Token.Literal
}

func (rs *ReturnStatment) String() string {
	var out strings.Builder
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(" = ")
	return out.String()
}

type ExpressionStatment struct {
	Token      token.Token // firs token of the expression
	Expression Expression
}

func (es *ExpressionStatment) statementNode() {}

func (es *ExpressionStatment) TokenLiteral() string {
	return es.Token.Literal
}

func (es *ExpressionStatment) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// Root node
type Program struct {
	Statements []Statement
}

func NewProgram() *Program {
	return &Program{Statements: []Statement{}}
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) == 0 {
		return ""
	}
	return p.Statements[0].TokenLiteral()
}

func (p *Program) String() string {
	var out strings.Builder
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}
