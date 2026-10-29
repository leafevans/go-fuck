package rules

import (
	"sync"
	"testing"

	"github.com/leafevans/go-fuck/internal/types"
)

func TestRegisterCustom(t *testing.T) {
	resetTestState(t)

	rule := types.NewRule("custom_test_rule",
		func(*types.Command) bool { return false },
		func(*types.Command) ([]string, error) { return nil, nil })
	if err := RegisterCustom(rule); err != nil {
		t.Fatalf("RegisterCustom() error = %v", err)
	}
	t.Cleanup(func() {
		registryMu.Lock()
		defer registryMu.Unlock()
		for index, registered := range registry {
			if registered.Name == rule.Name {
				registry = append(registry[:index], registry[index+1:]...)
				break
			}
		}
		delete(byName, rule.Name)
		allOnce = sync.Once{}
		allRules = nil
	})

	if got := Get(rule.Name); got != rule {
		t.Fatalf("Get(%q) = %v, want the custom rule", rule.Name, got)
	}
	found := false
	for _, registered := range All() {
		if registered.Name == rule.Name {
			found = true
		}
	}
	if !found {
		t.Fatalf("All() does not contain %q", rule.Name)
	}
	if err := RegisterCustom(rule); err == nil {
		t.Fatal("RegisterCustom() accepted a duplicate rule name")
	}
	if err := RegisterCustom(nil); err == nil {
		t.Fatal("RegisterCustom(nil) = nil, want an error")
	}
}
