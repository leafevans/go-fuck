package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run removes the given paths and returns the process exit code.
func run(paths []string, stdout, stderr io.Writer) int {
	if len(paths) == 0 {
		fmt.Fprintln(stderr, "clean: missing path argument")
		fmt.Fprintln(stderr, "usage: go run ./tools/clean PATH...")
		return 2
	}

	failed := false
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			fmt.Fprintf(stderr, "clean: %v\n", err)
			failed = true
			continue
		}
		if err := os.RemoveAll(path); err != nil {
			fmt.Fprintf(stderr, "clean: %v\n", err)
			failed = true
			continue
		}
		fmt.Fprintln(stdout, "removed", path)
	}
	if failed {
		return 1
	}
	return 0
}
