package opls

import (
	"fmt"
	"os"
	"strings"
)

// Command line flags
var (
	LongFormat  bool
	Recursive   bool
	ShowAll     bool
	ReverseSort bool
	SortByTime  bool
)

// ParseFlags parses command line flags and returns the remaining arguments
func ParseFlags() []string {
	var args []string
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-") && len(arg) > 1 && arg[1] != '-' {
			for _, ch := range arg[1:] {
				switch ch {
				case 'l':
					LongFormat = true
				case 'R':
					Recursive = true
				case 'a':
					ShowAll = true
				case 'r':
					ReverseSort = true
				case 't':
					SortByTime = true
				default:
					fmt.Fprintf(os.Stderr, "my-ls: invalid option -- '%c'\n", ch)
					os.Exit(1)
				}
			}
		} else {
			args = append(args, arg)
		}
	}
	if len(args) == 0 {
		args = []string{"."}
	}
	return args
}
