package rules

import (
	"errors"
	"fmt"
	"sync"

	"github.com/leafevans/go-fuck/internal/types"
)

// RegisterCustom adds a rule that was loaded from a user rule file.
//
// The compiled in rules are registered from init functions and never change.
// Custom rules are registered once per process before the first correction is
// computed, which is why the cached rule list is invalidated here. It returns
// an error when a rule with the same name is already registered.
func RegisterCustom(rule *types.Rule) error {
	if rule == nil {
		return errors.New("custom rule is nil")
	}

	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := byName[rule.Name]; exists {
		return fmt.Errorf("rule %q is already registered", rule.Name)
	}

	registry = append(registry, rule)
	byName[rule.Name] = rule

	// All caches its result in a sync.Once, so drop the cache. Registration
	// happens during startup, before corrections are computed.
	allOnce = sync.Once{}
	allRules = nil
	return nil
}
