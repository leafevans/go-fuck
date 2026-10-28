// Command help prints the documented targets of the Makefile.
//
// Make targets are documented with a trailing "## description" comment on
// their target line. Reading the Makefile with Go keeps the output readable
// and identical on every platform, without relying on grep, sed or the
// platform dependent behavior of shell echo built-ins.
//
// Usage:
//
//	go run ./tools/help
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// targetLine matches a Makefile target line that carries a help comment, for
// example "build: ## Build the go-fuck binary". The [^=]* part keeps the
// pattern from matching variable assignments such as "build := ...".
var targetLine = regexp.MustCompile(`^([A-Za-z0-9_-]+):[^=]*##\s+(.+)$`)

// target is one documented make target.
type target struct {
	name        string
	description string
}

func main() {
	makefile := flag.String("makefile", "Makefile", "path of the Makefile to inspect")
	flag.Parse()

	targets, err := parse(*makefile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "help:", err)
		os.Exit(1)
	}

	width := 0
	for _, t := range targets {
		if len(t.name) > width {
			width = len(t.name)
		}
	}

	fmt.Println("Usage: make TARGET")
	fmt.Println()
	fmt.Println("Targets:")
	for _, t := range targets {
		fmt.Printf("  %-*s  %s\n", width, t.name, t.description)
	}
}

// parse returns the documented targets of a Makefile, in file order.
func parse(path string) ([]target, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var targets []target
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		match := targetLine.FindStringSubmatch(scanner.Text())
		if match == nil {
			continue
		}
		targets = append(targets, target{
			name:        match[1],
			description: strings.TrimSpace(match[2]),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("no documented targets found in %s", path)
	}
	return targets, nil
}
