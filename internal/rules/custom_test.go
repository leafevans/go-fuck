package rules

import (
	"slices"
	"sync"
	"testing"

	"github.com/leafevans/go-fuck/v4/internal/types"
)

func TestRegisterCustom(t *testing.T) {
	resetTestState(t)

	rule := types.NewRule("custom_test_rule",
		func(*types.Command) bool { return false },
		func(*types.Command) ([]string, error) { return nil, nil })
	before := All()
	var readers sync.WaitGroup
	for range 4 {
		readers.Go(func() {
			for range 100 {
				_ = All()
				_ = Get(rule.Name)
			}
		})
	}
	err := RegisterCustom(rule)
	readers.Wait()
	if err != nil {
		t.Fatalf("RegisterCustom() error = %v", err)
	}
	t.Cleanup(func() {
		registryMu.Lock()
		defer registryMu.Unlock()
		delete(byName, rule.Name)
		orderedRules = nil
	})

	if got := Get(rule.Name); got != rule {
		t.Fatalf("Get(%q) = %v, want the custom rule", rule.Name, got)
	}
	if slices.Contains(before, rule) || len(All()) != len(before)+1 {
		t.Fatal("All() did not replace its snapshot after registration")
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
