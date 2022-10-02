package evaluator

import (
	"testing"

	"protiumx.dev/simia/lexer"
	"protiumx.dev/simia/parser"
	"protiumx.dev/simia/value"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerValue(t, evaluated, tt.expected)
	}
}

func testEval(input string) value.Value {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	return Eval(program)
}

func testIntegerValue(t *testing.T, val value.Value, expected int64) bool {
	result, ok := val.(*value.Integer)
	if !ok {
		t.Errorf("val is not integer. got=%T (%+v)", val, val)
		return false
	}
	if result.Value != expected {

		t.Errorf("val has wrong value. got=%d, expected=%d", result.Value, expected)
		return false
	}
	return true
}