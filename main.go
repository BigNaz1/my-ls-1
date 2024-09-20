package main

import (
	"fmt"
	"os"

	opls "opls/internal"
)

func main() {
	// Custom flag parsing
	args := opls.ParseFlags()

	// List files for each argument
	for i, arg := range args {
		if i > 0 {
			fmt.Println()
		}

		err := opls.HandleSingleFile(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "my-ls: cannot access '%s': %v\n", arg, err)
		}
	}
}
