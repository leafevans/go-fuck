// Package rules contains the go-fuck correction rules.
//
// Every rule lives in its own file next to this one and registers itself from
// an init function. The rule name normally matches the file name.
package rules

import (
	"sort"
	"sync"

	"github.com/leafevans/go-fuck/internal/types"
)

// Characters used by rules that build shell snippets or match quoting
// problems.
var (
	// newline is the line separator.
	newline = string(rune(10))
	// backslash is used to escape shell characters.
	backslash = string(rune(92))
	// doubleQuote is the double quote character.
	doubleQuote = string(rune(34))
)

var (
	registryMu sync.Mutex
	registry   []*types.Rule
	byName     = map[string]*types.Rule{}

	allOnce  sync.Once
	allRules []*types.Rule
)

// register adds a rule to the registry. It is called from the init function of
// each rule file.
func register(rule *types.Rule) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := byName[rule.Name]; exists {
		panic("go-fuck: duplicate rule " + rule.Name)
	}
	registry = append(registry, rule)
	byName[rule.Name] = rule
}

// Count returns the number of registered rules.
func Count() int {
	return len(All())
}

// All returns the registered rules ordered by priority and name.
//
// Rules are registered in file name order, which reproduces the stable sorting
// of the original implementation: rules are tried in priority order and
// alphabetically within the same priority.
//
// The list is computed once because the registry is immutable after the init
// phase. The returned slice is shared and must not be modified by callers.
func All() []*types.Rule {
	allOnce.Do(func() {
		registryMu.Lock()
		all := make([]*types.Rule, len(registry))
		copy(all, registry)
		registryMu.Unlock()

		sort.SliceStable(all, func(left, right int) bool {
			return all[left].Name < all[right].Name
		})
		sort.SliceStable(all, func(left, right int) bool {
			return all[left].Priority < all[right].Priority
		})
		allRules = all
	})
	return allRules
}

// Get returns the rule with the given name.
func Get(name string) *types.Rule {
	registryMu.Lock()
	defer registryMu.Unlock()
	return byName[name]
}
