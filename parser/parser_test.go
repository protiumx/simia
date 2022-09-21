package parser

import (
	"testing"

	"protiumx.dev/simia/ast"
	"protiumx.dev/simia/lexer"
)

func TestLetStatement(t *testing.T) {
	input := `
  let x = 5;
  let y = 10;
  let foo = 9999;
  `

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)
	if program == nil {
		t.Fatal("ParseProgram() returned nil")
	}
	if len(program.Statements) != 3 {
		t.Fatalf("invalid amount of statements. got=%d", len(program.Statements))
	}
	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foo"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
	}
}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("TokenLiteral is not 'let'. got=%s", s.TokenLiteral())
		return false
	}
	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("cannot cast statement as LetStatement. got=%T", s)
		return false
	}

	if letStmt.Name.Value != name {
		t.Errorf("incorrect name. expected=%s, got=%s", name, letStmt.Name.Value)
		return false
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("incorrect TokenLiteral. expected=%s, got=%s", name, letStmt.Name)
		return false
	}

	return true
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, err := range errors {
		t.Errorf("parser error: %q", err)
	}
	t.FailNow()
}

func TestReturnStatement(t *testing.T) {
	input := `
  return 5;
  return fn();
  return 999;
  `
	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d", len(program.Statements))
	}

	for _, stmt := range program.Statements {
		retStmt, ok := stmt.(*ast.ReturnStatment)
		if !ok {
			t.Errorf("cannot cast statement as *ast.ReturnStatment. got=%T", stmt)
			continue
		}
		if retStmt.TokenLiteral() != "return" {
			t.Errorf("incorrect token literal. got=%q", retStmt.TokenLiteral())
		}
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := "foobar;"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statements. got=%d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatment)
	if !ok {
		t.Fatalf("cannot cast program.Statements[0] as ast.ExpressionStatment. got=%T", program.Statements[0])
	}

	ident, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("cannot cast expression as ast.Identifier. got=%T", stmt.Expression)
	}

	if ident.Value != "foobar" {
		t.Errorf("wrong identifier value. expeted=foobar, got=%s", ident.Value)
	}

	if ident.TokenLiteral() != "foobar" {
		t.Errorf("wrong TokenLiteral value. expeted=foobar, got=%s", ident.TokenLiteral())
	}
}

func TestIntigerLiteralExpression(t *testing.T) {
	input := "0;"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statements. got=%d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatment)
	if !ok {
		t.Fatalf("cannot cast program.Statements[0] as ast.ExpressionStatment. got=%T", program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.IntigerLiteral)
	if !ok {
		t.Fatalf("cannot cast expression as ast.IntigerLiteral. got=%T", stmt.Expression)
	}

	if literal.Value != 0 {
		t.Errorf("wrong identifier value. expeted=0, got=%d", literal.Value)
	}

	if literal.TokenLiteral() != "0" {
		t.Errorf("wrong TokenLiteral value. expeted=0, got=%s", literal.TokenLiteral())
	}
}
