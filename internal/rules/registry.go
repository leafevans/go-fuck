// Package rules contains the go-fuck correction rules.
//
// Every rule lives in its own file next to this one and registers itself from
// an init function. The rule name normally matches the file name.
package rules

import (
	"cmp"
	"slices"
	"sync"

	"github.com/leafevans/go-fuck/v3/internal/specific"
	"github.com/leafevans/go-fuck/v3/internal/types"
	"github.com/leafevans/go-fuck/v3/internal/utils"
)

var startupTools = specific.ToolAvailability(utils.StartupAvailable)

// Characters used by rules that build shell snippets or match quoting
// problems.
var (
	// newline is the line separator.
	newline = "\n"
	// backslash is used to escape shell characters.
	backslash = "\\"
	// doubleQuote is the double quote character.
	doubleQuote = "\""
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
	registryMu.Lock()
	defer registryMu.Unlock()
	return len(registry)
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

		// Sorting by priority and then by name is equivalent to the previous
		// name sort followed by a stable priority sort, because the names are
		// unique.
		slices.SortFunc(all, func(left, right *types.Rule) int {
			if order := cmp.Compare(left.Priority, right.Priority); order != 0 {
				return order
			}
			return cmp.Compare(left.Name, right.Name)
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
