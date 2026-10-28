// Command gen-rule-list regenerates the rule list in README.md.
//
// The generated block is delimited by the "<!-- RULES_LIST_START -->" and
// "<!-- RULES_LIST_END -->" markers and contains one backticked rule name per
// line, sorted alphabetically. Implementing the rewrite in Go keeps it
// portable: a sed based script is not guaranteed to be available on Windows.
//
// Usage:
//
//	go run ./tools/gen-rule-list          # rewrite README.md
//	go run ./tools/gen-rule-list -check   # fail when README.md is out of date
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/rules"
)

const (
	// startMarker opens the generated rule list in README.md.
	startMarker = "<!-- RULES_LIST_START -->"
	// endMarker closes the generated rule list in README.md.
	endMarker = "<!-- RULES_LIST_END -->"
)

func main() {
	readme := flag.String("readme", "README.md", "path of the README file to update")
	check := flag.Bool("check", false, "report whether the rule list is up to date instead of rewriting it")
	flag.Parse()

	if err := run(*readme, *check); err != nil {
		fmt.Fprintln(os.Stderr, "gen-rule-list:", err)
		os.Exit(1)
	}
}

// run updates the rule list in one README file.
func run(path string, check bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	original := string(data)
	updated, err := rewrite(original, ruleNames())
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}

	if check {
		if updated != original {
			return fmt.Errorf("%s is out of date; run %q to regenerate it", path, "make rules-list")
		}
		fmt.Printf("%s: rule list is up to date\n", path)
		return nil
	}

	if updated == original {
		fmt.Printf("%s: rule list is already up to date\n", path)
		return nil
	}
	if err := os.WriteFile(path, []byte(updated), info.Mode().Perm()); err != nil {
		return err
	}
	fmt.Printf("%s: updated rule list\n", path)
	return nil
}

// ruleNames returns the names of all registered rules, sorted alphabetically.
func ruleNames() []string {
	conf.Reset()
	registered := rules.All()
	names := make([]string, 0, len(registered))
	for _, rule := range registered {
		names = append(names, rule.Name)
	}
	sort.Strings(names)
	return names
}

// rewrite replaces the lines between the rule list markers with a markdown
// list of the given rule names, preserving the line ending style and the
// marker lines of the document.
func rewrite(document string, names []string) (string, error) {
	eol := "\n"
	if strings.Contains(document, "\r\n") {
		eol = "\r\n"
	}
	lines := strings.Split(strings.ReplaceAll(document, "\r\n", "\n"), "\n")

	start, end := -1, -1
	for index, line := range lines {
		switch strings.TrimSpace(line) {
		case startMarker:
			if start != -1 {
				return "", fmt.Errorf("found more than one %q marker", startMarker)
			}
			start = index
		case endMarker:
			if end != -1 {
				return "", fmt.Errorf("found more than one %q marker", endMarker)
			}
			end = index
		}
	}
	switch {
	case start == -1:
		return "", fmt.Errorf("missing %q marker", startMarker)
	case end == -1:
		return "", fmt.Errorf("missing %q marker", endMarker)
	case end < start:
		return "", fmt.Errorf("%q marker appears before %q", endMarker, startMarker)
	}

	block := make([]string, 0, len(names)+2)
	block = append(block, lines[start])
	for _, name := range names {
		block = append(block, "- `"+name+"`")
	}
	block = append(block, lines[end])

	updated := make([]string, 0, len(lines))
	updated = append(updated, lines[:start]...)
	updated = append(updated, block...)
	updated = append(updated, lines[end+1:]...)

	result := strings.Join(updated, "\n")
	if eol == "\r\n" {
		result = strings.ReplaceAll(result, "\n", eol)
	}
	return result, nil
}
