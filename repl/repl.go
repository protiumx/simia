package repl

import (
	"bufio"
	"fmt"
	"io"

	"protiumx.dev/simia/compiler"
	"protiumx.dev/simia/lexer"
	"protiumx.dev/simia/parser"
	"protiumx.dev/simia/vm"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

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

		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			fmt.Fprintf(out, "compilation error:\n %s\n", err)
			continue
		}

		v := vm.New(comp.Bytecode())
		err = v.Run()
		if err != nil {
			fmt.Fprintf(out, "bytecode execution error:\n %s\n", err)
			continue
		}

		top := v.LastPoppedStackElement()
		io.WriteString(out, top.Inspect())
		io.WriteString(out, "\n")
	}
}

func printParseErrors(out io.Writer, errors []string) {
	io.WriteString(out, "parse errors:\n")
	for _, msg := range errors {
		io.WriteString(out, fmt.Sprintf("\t%s\n", msg))
	}
}
