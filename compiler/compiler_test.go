package compiler

import (
	"fmt"
	"testing"

	"protiumx.dev/simia/ast"
	"protiumx.dev/simia/code"
	"protiumx.dev/simia/lexer"
	"protiumx.dev/simia/parser"
	"protiumx.dev/simia/value"
)

type compilerTestcase struct {
	input                string
	expectedConstants    []any
	expectedInstructions []code.Instructions
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func runCompilerTests(t *testing.T, tests []compilerTestcase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.Bytecode()

		err = testInstructions(tt.expectedInstructions, bytecode.Instructions)
		if err != nil {
			t.Fatalf("test instructions failed: %s", err)
		}

		err = testConstants(t, tt.expectedConstants, bytecode.Constants)
		if err != nil {
			t.Fatalf("test constants failed: %s", err)
		}
	}
}

func testInstructions(expected []code.Instructions, actual code.Instructions) error {
	concatted := flattenInstructions(expected)

	if len(actual) != len(concatted) {
		return fmt.Errorf("wrong instructions length.\nwant=%q\ngot=%q", concatted, actual)
	}

	for i, ins := range concatted {
		if actual[i] != ins {
			return fmt.Errorf("wrong instruction at %d.\nwant\n%s\ngot\n%s", i, concatted, actual)
		}
	}

	return nil
}

func flattenInstructions(s []code.Instructions) code.Instructions {
	out := code.Instructions{}

	for _, ins := range s {
		out = append(out, ins...)
	}

	return out
}

func testConstants(t *testing.T, expected []any, actual []value.Value) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("wrong number of constants. got=%d, want=%d", len(actual), len(expected))
	}

	for i, constant := range expected {
		switch constant := constant.(type) {
		case int:
			err := testIntegerValue(int64(constant), actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testIntegerValue failed: %s", i, err)
			}
		case string:
			err := testStringValue(constant, actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testStringValue failed: %s", i, err)
			}
		case []code.Instructions:
			fn, ok := actual[i].(*value.CompiledFunction)
			if !ok {
				return fmt.Errorf("constant %d - not a function: %T", i, actual[i])
			}

			err := testInstructions(constant, fn.Instructions)
			if err != nil {
				return fmt.Errorf("constant %d - testInstruction failed: %s", i, err)
			}
		}
	}

	return nil
}

func testStringValue(expected string, actual value.Value) error {
	result, ok := actual.(*value.String)
	if !ok {
		return fmt.Errorf("value is not String. got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("value has wrong value. got=%q, want=%q", result.Value, expected)
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

func TestIntegerArithmetic(t *testing.T) {
	tests := []compilerTestcase{
		{
			"1 + 2",
			[]any{1, 2},
			[]code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},

		{
			"1; 2",
			[]any{1, 2},
			[]code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			"1 - 2",
			[]any{1, 2},
			[]code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSub),
				code.Make(code.OpPop),
			},
		},
		{
			"1 * 2",
			[]any{1, 2},
			[]code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpMul),
				code.Make(code.OpPop),
			},
		},
		{
			"1 / 2",
			[]any{1, 2},
			[]code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpDiv),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "-1",
			expectedConstants: []any{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpMinus),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []compilerTestcase{
		{
			input:             "true",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "false",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpFalse),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 > 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreaterThan),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 < 2",
			expectedConstants: []any{2, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreaterThan),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 == 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 != 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpNotEqual),
				code.Make(code.OpPop),
			},
		},

		{
			input:             "!true",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpBang),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []compilerTestcase{
		{
			input: `
      if true { 10 }; 11;
      `,
			expectedConstants: []any{10, 11},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpJumpIfBranch, 10),
				code.Make(code.OpConstant, 0),
				code.Make(code.OpJump, 11),
				code.Make(code.OpNil),
				// 0011
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
      if true { 10 } else { 20 }; 11;
      `,
			expectedConstants: []any{10, 20, 11},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpJumpIfBranch, 10),
				code.Make(code.OpConstant, 0),
				// 0007
				code.Make(code.OpJump, 13),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestGlobalLetStatements(t *testing.T) {
	tests := []compilerTestcase{
		{
			input: `
      let one = 1;
      let two = 2;
      `,
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSetGlobal, 1),
			},
		},
		{
			input: `
      let one = 1;
      one;
      `,
			expectedConstants: []any{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []compilerTestcase{
		{
			input:             `"monkey"`,
			expectedConstants: []any{"monkey"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},

		{
			input:             `"mon" + "key"`,
			expectedConstants: []any{"mon", "key"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []compilerTestcase{
		{
			input:             "[]",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpArray, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "[1, 3 - 4, 5 * 6]",
			expectedConstants: []any{1, 3, 4, 5, 6},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpSub),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpMul),
				code.Make(code.OpArray, 3), // 3 elements in the array
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestHashLiterals(t *testing.T) {
	tests := []compilerTestcase{
		{
			input:             "{}",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpHash, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             `{"a": 2, "b": 3}`,
			expectedConstants: []any{"a", 2, "b", 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpHash, 4),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []compilerTestcase{
		{
			input:             "[1,2,3][1+1]",
			expectedConstants: []any{1, 2, 3, 1, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpAdd),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
		{
			input:             `{"a": 1}["a"]`,
			expectedConstants: []any{"a", 1, "a"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpHash, 2),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestFunctions(t *testing.T) {
	tests := []compilerTestcase{
		{
			input: `fn() { return 1 + 3 }`,
			expectedConstants: []any{
				1,
				3,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
		},
		{
			input: `fn() { 1 + 3 }`,
			expectedConstants: []any{
				1,
				3,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
		},
		{
			input: `fn() { 1; 2 }`,
			expectedConstants: []any{
				1,
				2,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpPop),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
		},
		{
			input: `fn() { }`,
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestScopes(t *testing.T) {
	compiler := New()
	if compiler.scopeIndex != 0 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d", compiler.scopeIndex, 0)
	}

	compiler.emit(code.OpMul)
	compiler.enterScope()

	if compiler.scopeIndex != 1 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d", compiler.scopeIndex, 1)
	}

	compiler.emit(code.OpSub)
	if l := len(compiler.scopes[compiler.scopeIndex].instructions); l != 1 {
		t.Errorf("instructions length wrong. got=%d, want=%d", l, 1)
	}

	last := compiler.scopes[compiler.scopeIndex].lastInstruction
	if last.Opcode != code.OpSub {
		t.Errorf("lastInstruction.Opcode wrong. got=%d, want=%d", last.Opcode, code.OpSub)
	}

	compiler.leaveScope()
	if compiler.scopeIndex != 0 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d", compiler.scopeIndex, 0)
	}

	compiler.emit(code.OpAdd)
	if l := len(compiler.scopes[compiler.scopeIndex].instructions); l != 2 {
		t.Errorf("instructions length wrong. got=%d, want=%d", l, 2)
	}

	last = compiler.scopes[compiler.scopeIndex].lastInstruction
	if last.Opcode != code.OpAdd {
		t.Errorf("lastInstruction.Opcode wrong. got=%d, want=%d", last.Opcode, code.OpAdd)
	}

	prev := compiler.scopes[compiler.scopeIndex].prevInstruction
	if prev.Opcode != code.OpMul {
		t.Errorf("prevInstruction.Opcode wrong. got=%d, want=%d", last.Opcode, code.OpMul)
	}
}

func TestFunctionCalls(t *testing.T) {
  tests := []compilerTestcase{
    {
      input: `fn() { 24 }()`,
      expectedConstants: []any{
        24,
        []code.Instructions{
          code.Make(code.OpConstant, 0),
          code.Make(code.OpReturnValue),
        },
      },
      expectedInstructions: []code.Instructions{
        code.Make(code.OpConstant, 1),
        code.Make(code.OpCall),
        code.Make(code.OpPop),
      },
    },
    {
      input: `
      let f = fn() { 2 };
      f();
      `,
      expectedConstants: []any{
        2,
        []code.Instructions{
          code.Make(code.OpConstant, 0),
          code.Make(code.OpReturnValue),
        },
      },
      expectedInstructions: []code.Instructions{
        code.Make(code.OpConstant, 1),
        code.Make(code.OpSetGlobal, 0),
        code.Make(code.OpGetGlobal, 0),
        code.Make(code.OpCall),
        code.Make(code.OpPop),
      },
    },
  }

  runCompilerTests(t, tests)
}
