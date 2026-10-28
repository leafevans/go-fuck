// Command clean removes the build artifacts of go-fuck.
//
// It is used by "make clean" and exists because "rm -rf" is not available in
// every Windows shell. Missing paths are ignored.
//
// Usage:
//
//	go run ./tools/clean PATH...
package main

import (
	"fmt"
	"os"
)

func main() {
	paths := os.Args[1:]
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "clean: missing path argument")
		fmt.Fprintln(os.Stderr, "usage: go run ./tools/clean PATH...")
		os.Exit(2)
	}

	failed := false
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			fmt.Fprintf(os.Stderr, "clean: %v\n", err)
			failed = true
			continue
		}
		if err := os.RemoveAll(path); err != nil {
			fmt.Fprintf(os.Stderr, "clean: %v\n", err)
			failed = true
			continue
		}
		fmt.Println("removed", path)
	}
	if failed {
		os.Exit(1)
	}
}
