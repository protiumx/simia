package repl

import (
	"bufio"
	"fmt"
	"io"

	"protiumx.dev/simia/evaluator"
	"protiumx.dev/simia/lexer"
	"protiumx.dev/simia/parser"
	"protiumx.dev/simia/value"
)

const PROMPT = ">>"

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := value.NewEnvironment()

	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			printParseErrors(out, p.Errors())
			continue
		}

		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

func printParseErrors(out io.Writer, errors []string) {
	io.WriteString(out, "parse errors:\n")
	for _, msg := range errors {
		io.WriteString(out, fmt.Sprintf("\t%s\n", msg))
	}
}
