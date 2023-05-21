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

func runVMTests(t *testing.T, tests []vmTestCase) {
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

	case map[string]int64:
		h, ok := actual.(*value.Hash)
		if !ok {
			t.Errorf("value is not Hash. got=%T (%+v)", actual, actual)
			return
		}
		if len(h.Pairs) != len(expected) {
			t.Errorf("hash has wrong number of pairs. want=%d, got=%d", len(expected), len(h.Pairs))
			return
		}

		for k, v := range expected {
			err := testIntegerValue(v, h.Pairs[k])
			if err != nil {
				t.Errorf("testIntegerValue failed: %s", err)
			}
		}

	case *value.Nil:
		if actual != Nil {
			t.Errorf("test nil is not Nil: %T (%+v)", actual, actual)
		}

	case *value.Error:
		errValue, ok := actual.(*value.Error)
		if !ok {
			t.Errorf("value is not Error: %T (%+v)", actual, actual)
			return
		}

		if errValue.Message != expected.Message {
			t.Errorf("wrong error message. expected=%q, got=%q", expected.Message, errValue.Message)
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

	runVMTests(t, tests)
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

	runVMTests(t, tests)
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

	runVMTests(t, tests)
}

func TestGlobalLetStatements(t *testing.T) {
	tests := []vmTestCase{
		{"let one = 1; one;", 1},
		{"let one = 1; let two = 2; one + two;", 3},
	}

	runVMTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []vmTestCase{
		{`"monkey"`, "monkey"},
		{`"mon" + "key"`, "monkey"},
	}

	runVMTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []vmTestCase{
		{"[]", []int{}},
		{"[1 - 2, 3 * 4]", []int{-1, 12}},
	}

	runVMTests(t, tests)
}

func TestHashListerals(t *testing.T) {
	tests := []vmTestCase{
		{
			"{}", map[string]int64{},
		},
		{

			`{"a": 1 * 2, "b": 2+6}`,
			map[string]int64{
				"a": 2,
				"b": 8,
			},
		},
	}

	runVMTests(t, tests)
}

func TestIndexExpression(t *testing.T) {
	tests := []vmTestCase{
		{"[1, 2, 3][1]", 2},
		{"[1, 2, 3][0 + 2]", 3},
		{"[[1, 2], 3][0][0]", 1},
		{"[][0]", Nil},
		{`{"a": false }["a"]`, False},
		{`{}["a"]`, Nil},
	}

	runVMTests(t, tests)
}

func TestCallingFunction(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
      let sumFive = fn() { 4 + 1 };
      sumFive();
      `,
			expected: 5,
		},
		{
			input: `
      let five = fn() { 4 + 1 };
      let a = fn() { five() + 1};
      let b = fn() { a() + 1 };
      b();
      `,
			expected: 7,
		},
		{
			input: `
      let earlyReturn = fn() { return 1; 100; };
      earlyReturn();
      `,
			expected: 1,
		},
		{
			input: `
      let noReturn = fn() { };
      noReturn()
      `,
			expected: Nil,
		},
		{
			input: `
      let a = fn() { 1; };
      let b = fn() { a; };
      b()()
      `,
			expected: 1,
		},
	}

	runVMTests(t, tests)
}

func TestCallingFunctionsWithBindings(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let one = fn() { let one = 1; one };
			one();
			`,
			expected: 1,
		},
		{
			input: `
			let test = fn() {
				let a = 1;
				let b = 2;
				a + b
			};
			test();
			`,
			expected: 3,
		},
		{
			input: `
			let a = fn() { let foo = 10; foo; }
			let b = fn() { let foo = 100; foo; }
			a() + b();
			`,
			expected: 110,
		},
		{
			input: `
			let global = 100;
			let a = fn() {
				let num = 1;
				global - num;
			}
			let b = fn() {
				let num = 2;
				global - num;
			}
			a() + b();
			`,
			expected: 197,
		},
		{
			input: `
			let a = fn() {
				let b = fn () { 1; }
				b;
			}
			a()()
			`,
			expected: 1,
		},
	}
	runVMTests(t, tests)
}

func TestCallingFunctionsWithArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let identity = fn(a) { a };
			identity(0);
			`,
			expected: 0,
		},
		{
			input: `
			let sum = fn(a, b) { a + b };
			sum(1, 2);
			`,
			expected: 3,
		},
		{
			input: `
			let sum = fn(a, b) {
				let c = a + b;
				c;
			}
			sum(1, 2);
			`,
			expected: 3,
		},
		{
			input: `
			let global = 100;

			let sum = fn(a, b) {
				let c = a + b;
				c + global;
			}

			let test = fn() {
				sum(1,2) + sum(3,4);
			}
			test();
			`,
			expected: 210,
		},
	}

	runVMTests(t, tests)
}

func TestFunctionCallErrors(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    `fn() {}(1);`,
			expected: "wrong number of arguments: want=0, got=1",
		},
		{
			input:    `fn(a, b) {}(1);`,
			expected: "wrong number of arguments: want=2, got=1",
		},
	}

	for _, tt := range tests {
		program := parse(tt.input)
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(comp.Bytecode())
		err = vm.Run()
		if err == nil {
			t.Fatalf("expected VM error")
		}

		if err.Error() != tt.expected {
			t.Fatalf("wrong VM error. want=%q, got=%q", tt.expected, err)
		}
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []vmTestCase{
		{`len("")`, 0},
		{`len("1234")`, 4},
		{`len(0)`, &value.Error{Message: "argument to `len` not supported, got INTEGER"}},
		{`len(1,2)`, &value.Error{Message: "wrong number of arguments. got=2, want=1"}},
		{`len([1])`, 1},
		{`log("test")`, Nil},
		{`append([], 1)`, []int{1}},
		{`append(1, 1)`, &value.Error{Message: "argument must be ARRAY, got INTEGER"}},
	}

	runVMTests(t, tests)
}
