package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args

	env, err := ReadDir(args[1])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		exitCode := RunCmd(args[2:], env)
		os.Exit(exitCode)
	}
}
