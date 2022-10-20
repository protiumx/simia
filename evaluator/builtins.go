package evaluator

import "protiumx.dev/simia/value"

var builtins = map[string]*value.Builtin{
	"len": {
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *value.String:
				return &value.Integer{Value: int64(len(arg.Value))}
			default:
				return newError("argument to `len` not supported, got %s", arg.Type())
			}
		},
	},
}
