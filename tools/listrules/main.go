// Command listrules prints the names of all correction rules compiled into
// go-fuck, one per line and sorted alphabetically.
//
// The rule list in README.md is generated from this output by
// "go run ./tools/gen-rule-list". The command is also useful to discover the
// rule names accepted by the "rules", "exclude_rules" and "priority" settings.
//
// Usage:
//
//	go run ./tools/listrules
package main

import (
	"fmt"
	"sort"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/rules"
)

func main() {
	for _, name := range ruleNames() {
		fmt.Println(name)
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
