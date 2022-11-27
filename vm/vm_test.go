package vm

import (
	"fmt"
	"testing"

	"protiumx.dev/simia/ast"
	"protiumx.dev/simia/compiler"
	"protiumx.dev/simia/lexer"
	"protiumx.dev/simia/parser"
	"protiumx.dev/simia/value"
)

type vmTestCase struct {
	input    string
	expected any
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(comp.Bytecode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm erro: %s", err)
		}

		stackElement := vm.LastPoppedStackElement()

		testExpectedValue(t, tt.expected, stackElement)
	}
}

func testExpectedValue(t *testing.T, expected any, actual value.Value) {
	t.Helper()

	switch expected := expected.(type) {
	case int:
		err := testIntegerValue(int64(expected), actual)
		if err != nil {
			t.Errorf("test integer value failed: %s", err)
		}
	case bool:
		err := testBooleanValue(bool(expected), actual)
		if err != nil {
			t.Errorf("test bool value failed: %s", err)
		}
	}
}

func testIntegerValue(expected int64, actual value.Value) error {
	result, ok := actual.(*value.Integer)
	if !ok {
		return fmt.Errorf("value is not Integer. got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("value has wrong value. got=%d, want=%d", result.Value, expected)
	}

	return nil
}

func testBooleanValue(expected bool, actual value.Value) error {
	result, ok := actual.(*value.Boolean)
	if !ok {
		return fmt.Errorf("value is not Boolean. got=%T (%+v)", actual, actual)
	}
	if result.Value != expected {
		return fmt.Errorf("value has wrong value. got=%t, want=%t", result.Value, expected)
	}

	return nil
}

func TestIntegerAithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1", 1},
		{"1 + 2", 3},
		{"1 - 2", -1},
		{"3 * 2 - (6 / 3)", 4},
		{"1 * 2 - 3 / 1", -1},
	}

	runVmTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"true", true},
		{"false", false},
	}

	runVmTests(t, tests)
}
