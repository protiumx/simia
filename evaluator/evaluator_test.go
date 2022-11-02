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
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
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
	env := value.NewEnvironment(nil)
	return Eval(program, env)
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

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanValue(t, evaluated, tt.expected)
	}
}

func testBooleanValue(t *testing.T, val value.Value, expected bool) bool {
	result, ok := val.(*value.Boolean)
	if !ok {
		t.Errorf("val is not Boolean. got=%T (%+v)", val, val)
		return false
	}

	if result.Value != expected {
		t.Errorf("val has wrong value. got=%t, expected=%t", result.Value, expected)
		return false
	}
	return true
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanValue(t, evaluated, tt.expected)
	}
}

func TestIfElseExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerValue(t, evaluated, int64(integer))
		} else {
			testNilValue(t, evaluated)
		}
	}
}

func testNilValue(t *testing.T, val value.Value) bool {
	if val != NIL {
		t.Errorf("value is not NIL. got=%T (%+v)", val, val)
		return false
	}
	return true
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{`
    if (10 > 1) { 
      if (10 > 2) { 
        return 10; 
      } 
    }
    return 1;
    `, 10},
		{
			`
    let f = fn(x) {
      return x;
      x + 10;
    };
    f(10);`,
			10,
		},
		{
			`
    let f = fn(x) {
       let result = x + 10;
       return result;
       return 10;
    };
    f(10);`,
			20,
		},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerValue(t, evaluated, tt.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{
			"5 + true;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5 + true; 5;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-true",
			"unknown operator: -BOOLEAN",
		},
		{
			"true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"6; true + false; 7",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"if (10 > 1) { true + false; }",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`
    if (10 > 1) {
      if (10 > 1) {
        return true + false;
      }
    return 1; }
    `,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"foo;",
			"identifier not found: foo",
		},
		{
			`"simia" - "sim"`,
			"unknown operator: STRING - STRING",
		},
		{
			`{"name": "test"}[fn() {}];`,
			"key is not string: FN",
		},
		{
			`7 |> 0;`,
			"expected FUNCTION in pipiline expression. got=*ast.IntegerLiteral",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		err, ok := evaluated.(*value.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T(%+v)", evaluated, evaluated)
			continue
		}
		if err.Message != tt.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expectedMessage, err.Message)
		}
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
	}
	for _, tt := range tests {
		testIntegerValue(t, testEval(tt.input), tt.expected)
	}
}

func TestFunctionValue(t *testing.T) {
	input := "fn(x) { x + 2; }"

	evalualed := testEval(input)
	fn, ok := evalualed.(*value.Function)
	if !ok {
		t.Fatalf("value is not a Funtion. got=%T (%+v)", evalualed, evalualed)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong number of parameters. params=%+v", fn.Parameters)
	}

	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", fn.Parameters[0])
	}

	expectedBody := "(x + 2)"
	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got=%q", expectedBody, fn.Body.String())
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let identity = fn(x) { x; }; identity(5);", 5},
		{"let identity = fn(x) { return x; }; identity(5);", 5},
		{"let double = fn(x) { 2 * x; }; double(10);", 20},
		{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5, add(5, 5));", 15},
		{"fn(x) { x; }(5);", 5},
	}

	for _, tt := range tests {
		testIntegerValue(t, testEval(tt.input), tt.expected)
	}
}

func TestFunctionPipeLine(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let double = fn(x) { x * 2 }; 3 |> double();", 6},
		{"let add = fn(x, y) { x + y }; 3 |> add(7);", 10},
		{"let double = fn(x) { x * 2 }; let add = fn(x, y) { x + y }; 3 |> add(7) |> double();", 20},
		{"let add = fn(x, y) { x + y }; 3 |> add(add(1, 1));", 5},
	}

	for _, tt := range tests {
		testIntegerValue(t, testEval(tt.input), tt.expected)
	}
}

func TestClosures(t *testing.T) {
	input := `
  let newAdder = fn(x) {
    fn(y) { x + y; };
  };
  let addTwo = newAdder(2);
  addTwo(2);
  `
	testIntegerValue(t, testEval(input), 4)
}

func TestStringLiteral(t *testing.T) {
	input := `"test dev"`
	evaluated := testEval(input)
	str, ok := evaluated.(*value.String)
	if !ok {
		t.Fatalf("value is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "test dev" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestStringConcatenation(t *testing.T) {
	input := `"simia" + " " + "lang"`
	evaluated := testEval(input)
	str, ok := evaluated.(*value.String)
	if !ok {
		t.Fatalf("value is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "simia lang" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len(1)`, "argument to `len` not supported, got INTEGER"},
		{`len("one", "two")`, "wrong number of arguments. got=2, want=1"},
		{`len([1, 2, 3])`, 3},
		{`len([])`, 0},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerValue(t, evaluated, int64(expected))
		case string:
			err, ok := evaluated.(*value.Error)
			if !ok {
				t.Errorf("value is not Error. got=%T (%+v)", evaluated, evaluated)
			}

			if err.Message != expected {
				t.Errorf("wrong error message. expected=%q, got=%q", expected, err.Message)
			}
		}
	}
}

func TestArrayLiteral(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	evaluated := testEval(input)
	result, ok := evaluated.(*value.Array)
	if !ok {
		t.Fatalf("value is not Array. got=%T (%+v)", evaluated, evaluated)
	}

	if len(result.Elements) != 3 {
		t.Fatalf("array has wrong number of elements. got=%d", len(result.Elements))
	}

	testIntegerValue(t, result.Elements[0], 1)
	testIntegerValue(t, result.Elements[1], 4)
	testIntegerValue(t, result.Elements[2], 6)
}

func TestArrayIndexExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{
			"[1, 2, 3][0]",
			1,
		},
		{
			"let i = 0; [1][i];",
			1,
		},
		{
			"[1, 2][3 - 2];",
			2,
		},
		{
			"let arr = [1, 2, 3]; arr[2]",
			3,
		},
		{
			"let arr = [1, 2, 3]; arr[0] + arr[1];",
			3,
		},
		{
			"let arr = [1, 2, 3]; let i = arr[0]; arr[i];",
			2,
		},
		{
			"[1, 2][3];",
			nil,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerValue(t, evaluated, int64(integer))
		} else {
			testNilValue(t, evaluated)
		}
	}
}

func TestHashLiteral(t *testing.T) {
	input := `
  let a = "a";
  {
    "b": 10 - 1,
    "c" + "d": 6,
    a: 1
  }
  `
	evaluated := testEval(input)
	resutl, ok := evaluated.(*value.Hash)
	if !ok {
		t.Fatalf("Eval result is not a Hash. got=%T (%+v)", evaluated, evaluated)
	}

	expected := map[string]int64{
		"b":  9,
		"cd": 6,
		"a":  1,
	}
	if len(resutl.Pairs) != len(expected) {
		t.Fatalf("Hash has wrong num of pairs. got=%d", len(resutl.Pairs))
	}

	for k, v := range expected {
		val, ok := resutl.Pairs[k]
		if !ok {
			t.Errorf("no pair for key %s", k)
		}

		testIntegerValue(t, val, v)
	}
}

func TestHasIndexExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{
			`{"foo": 5}["foo"]`,
			5,
		},
		{
			`{"foo": 5}["test"]`,
			nil,
		},
		{
			`let foo = "foo"; {"foo": 5}[foo]`,
			5,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		val, ok := tt.expected.(int)
		if !ok {
			testNilValue(t, evaluated)
		} else {
			testIntegerValue(t, evaluated, int64(val))
		}
	}
}
