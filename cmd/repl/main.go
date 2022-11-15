package main

import (
	"fmt"
	"os"

	"protiumx.dev/simia/repl"
)

var Version = ""

func main() {
	fmt.Printf("simia %s\n", Version)
	repl.Start(os.Stdin, os.Stdout)
}
