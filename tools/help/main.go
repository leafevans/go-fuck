package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
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
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run prints the documented Makefile targets and returns the process exit
// code.
func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("help", flag.ContinueOnError)
	flags.SetOutput(stderr)
	makefile := flags.String("makefile", "Makefile", "path of the Makefile to inspect")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	targets, err := parse(*makefile)
	if err != nil {
		fmt.Fprintln(stderr, "help:", err)
		return 1
	}

	width := 0
	for _, t := range targets {
		if len(t.name) > width {
			width = len(t.name)
		}
	}

	fmt.Fprintln(stdout, "Usage: make TARGET")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Targets:")
	for _, t := range targets {
		fmt.Fprintf(stdout, "  %-*s  %s", width, t.name, t.description)
		fmt.Fprintln(stdout)
	}
	return 0
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
		return nil, fmt.Errorf("no documented targets in %s", path)
	}
	return targets, nil
}
