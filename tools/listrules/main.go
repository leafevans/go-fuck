package main

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/leafevans/go-fuck/v2/internal/conf"
	"github.com/leafevans/go-fuck/v2/internal/rules"
)

func main() {
	printRules(os.Stdout)
}

// printRules writes the sorted rule names to w.
func printRules(w io.Writer) {
	for _, name := range ruleNames() {
		fmt.Fprintln(w, name)
	}
}

// ruleNames returns the sorted names of all registered rules.
func ruleNames() []string {
	// Rule registration does not depend on the settings, but loading the
	// defaults keeps the output identical to what the real command sees.
	conf.Reset()

	registered := rules.All()
	names := make([]string, 0, len(registered))
	for _, rule := range registered {
		names = append(names, rule.Name)
	}
	sort.Strings(names)
	return names
}
