package main

import (
	"syscall/js"

	"protiumx.dev/simia/evaluator"
	"protiumx.dev/simia/lexer"
	"protiumx.dev/simia/parser"
	"protiumx.dev/simia/value"
)

var Version = ""

func wrapper(env *value.Environment) js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 1 {
			return "ERROR: invalid arguments"
		}

		code := args[0].String()
		l := lexer.New(code)
		p := parser.New(l)
		program := p.ParseProgram()

		if errs := p.Errors(); len(errs) != 0 {
			return errs[len(errs)-1]
		}

		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			return evaluated.Inspect()
		}

		return ""
	})
	return jsonFunc
}

func main() {
	env := value.NewEnvironment(nil)
	js.Global().Set("simia", wrapper(env))
	js.Global().Set("simia_version", js.ValueOf(Version))
	<-make(chan struct{})
}
