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
	case string:
		err := testStringValue(expected, actual)
		if err != nil {
			t.Errorf("test string value failed: %s", err)
		}
	case []int:
		arr, ok := actual.(*value.Array)
		if !ok {
			t.Errorf("value is not Array: %T (%+v)", actual, actual)
			return
		}

		if len(arr.Elements) != len(expected) {
			t.Errorf("wrong num of elements. want=%d, got=%d", len(expected), len(arr.Elements))
			return
		}

		for i, expElement := range expected {
			err := testIntegerValue(int64(expElement), arr.Elements[i])
			if err != nil {
				t.Errorf("testIntegerValue failed: %s", err)
			}
		}

	case *value.Nil:
		if actual != Nil {
			t.Errorf("test nil is not Nil: %T (%+v)", actual, actual)
		}
	}
}

func testStringValue(expected string, actual value.Value) error {
	s, ok := actual.(*value.String)
	if !ok {
		return fmt.Errorf("value is not String. got=%T (%+v)", actual, actual)
	}

	if s.Value != expected {
		return fmt.Errorf("string has wrong value. got=%q, want=%q", s.Value, expected)
	}

	return nil
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
		{"-50 + 100 + -50", 0},
	}

	runVmTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"false == false", true},
		{"(1 < 2) == true", true},
		{"!true", false},
		{"!!false", false},
	}

	runVmTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []vmTestCase{
		{"if true { 10 }", 10},
		{"if (true) { 10 } else { 20 }", 10},
		{"if false { 10 } else { 20 }", 20},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if false { 10 }", Nil},
		{"if (if false { 10 }) { 10 } else { 20 }", 20},
	}

	runVmTests(t, tests)
}

func TestGlobalLetStatements(t *testing.T) {
	tests := []vmTestCase{
		{"let one = 1; one;", 1},
		{"let one = 1; let two = 2; one + two;", 3},
	}

	runVmTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []vmTestCase{
		{`"monkey"`, "monkey"},
		{`"mon" + "key"`, "monkey"},
	}

	runVmTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []vmTestCase{
		{"[]", []int{}},
		{"[1 - 2, 3 * 4]", []int{-1, 12}},
	}

	runVmTests(t, tests)
}
