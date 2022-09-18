package main

import (
	"fmt"
	"os"
	"os/user"

	"protiumx.dev/simia/repl"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	fmt.Printf("User %s\n", user.Username)
	repl.Start(os.Stdin, os.Stdout)
}
