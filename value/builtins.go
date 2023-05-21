package value

import "fmt"

var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{
		Name: "len",
		Builtin: &Builtin{
			Fn: func(args ...Value) Value {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1", len(args))
				}

				switch arg := args[0].(type) {
				case *String:
					return &Integer{Value: int64(len(arg.Value))}
				case *Array:
					return &Integer{Value: int64(len(arg.Elements))}
				default:
					return newError("argument to `len` not supported, got %s", arg.Type())
				}
			},
		},
	},
	{
		Name: "log",
		Builtin: &Builtin{
			Fn: func(args ...Value) Value {
				for _, arg := range args {
					fmt.Println(arg.Inspect())
				}
				return nil
			},
		},
	},

	{
		Name: "append",
		Builtin: &Builtin{
			Fn: func(args ...Value) Value {
				if len(args) != 2 {
					return newError("wrong number of arguments. got=%d, want=2", len(args))
				}

				if t := args[0].Type(); t != ARRAY_VALUE {
					return newError("argument must be %s, got %s", ARRAY_VALUE, t)
				}

				arr := args[0].(*Array)
				length := len(arr.Elements)

				newElements := make([]Value, length+1, length+1)
				copy(newElements, arr.Elements)
				newElements[length] = args[1]

				return &Array{Elements: newElements}
			},
		},
	},
}

func GetBuiltinByName(name string) *Builtin {
	for _, b := range Builtins {
		if b.Name == name {
			return b.Builtin
		}
	}

	return nil
}

func newError(format string, args ...any) *Error {
	return &Error{Message: fmt.Sprintf(format, args...)}
}
