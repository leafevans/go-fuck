// Package rules contains the go-fuck correction rules.
//
// Every rule lives in its own file next to this one and registers itself from
// an init function. The rule name normally matches the file name.
package rules

import (
	"cmp"
	"slices"
	"sync"

	"github.com/leafevans/go-fuck/v4/internal/specific"
	"github.com/leafevans/go-fuck/v4/internal/types"
	"github.com/leafevans/go-fuck/v4/internal/utils"
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
	registryMu   sync.Mutex
	byName       = map[string]*types.Rule{}
	orderedRules []*types.Rule
)

// register is used by the built-in rule initializers.
func register(rule *types.Rule) {
	if !addRule(rule) {
		panic("go-fuck: duplicate rule " + rule.Name)
	}
}

// addRule updates the name index and invalidates the ordered snapshot together.
func addRule(rule *types.Rule) bool {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := byName[rule.Name]; exists {
		return false
	}
	byName[rule.Name] = rule
	orderedRules = nil
	return true
}

// Count returns the number of registered rules.
func Count() int {
	registryMu.Lock()
	defer registryMu.Unlock()
	return len(byName)
}

// All returns a shared, read-only snapshot ordered by priority and name.
// Later registrations do not change snapshots already returned to callers.
func All() []*types.Rule {
	registryMu.Lock()
	defer registryMu.Unlock()
	if orderedRules == nil {
		orderedRules = make([]*types.Rule, 0, len(byName))
		for _, rule := range byName {
			orderedRules = append(orderedRules, rule)
		}
		slices.SortFunc(orderedRules, func(left, right *types.Rule) int {
			if order := cmp.Compare(left.Priority, right.Priority); order != 0 {
				return order
			}
			return cmp.Compare(left.Name, right.Name)
		})
	}
	return orderedRules
}

// Get returns the rule with the given name.
func Get(name string) *types.Rule {
	registryMu.Lock()
	defer registryMu.Unlock()
	return byName[name]
}
