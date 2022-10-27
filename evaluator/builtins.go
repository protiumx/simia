package evaluator

import (
	"fmt"

	"protiumx.dev/simia/value"
)

var builtins = map[string]*value.Builtin{
	"len": {
		Fn: func(args ...value.Value) value.Value {
			if err := checkArgsNumber(1, args); err != nil {
				return err
			}

			switch arg := args[0].(type) {
			case *value.String:
				return &value.Integer{Value: int64(len(arg.Value))}
			case *value.Array:
				return &value.Integer{Value: int64(len(arg.Elements))}
			default:
				return newError("argument to `len` not supported, got %s", arg.Type())
			}
		},
	},
	"append": {
		Fn: func(args ...value.Value) value.Value {
			if err := checkArgsNumber(2, args); err != nil {
				return err
			}

			if err := checkArgType(args[0].Type(), value.ARRAY_VALUE, "append"); err != nil {
				return err
			}

			arr := args[0].(*value.Array)
			length := len(arr.Elements)

			newElements := make([]value.Value, length+1, length+1)
			copy(newElements, arr.Elements)
			newElements[length] = args[1]

			return &value.Array{Elements: newElements}
		},
	},
	"log": {
		Fn: func(args ...value.Value) value.Value {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return NIL
		},
	},
}

func checkArgsNumber(n int, args []value.Value) *value.Error {
	if len(args) != n {
		return newError("wrong number of arguments. got=%d, want=%d", len(args), n)
	}

	return nil
}

func checkArgType(argType value.ValueType, expectedType value.ValueType, fnName string) *value.Error {
	if argType != expectedType {
		return newError("argument to `%s` must be %s, got %s", fnName, argType, expectedType)
	}

	return nil
}
