package main

import (
	"sort"
	"testing"

	"github.com/leafevans/go-fuck/internal/rules"
)

func TestRuleNames(t *testing.T) {
	names := ruleNames()
	if len(names) != rules.Count() {
		t.Fatalf("ruleNames() returned %d names, want %d", len(names), rules.Count())
	}
	if !sort.StringsAreSorted(names) {
		t.Fatalf("ruleNames() is not sorted")
	}
	found := false
	for _, name := range names {
		if name == "git_push" {
			found = true
		}
	}
	if !found {
		t.Fatalf("ruleNames() does not contain git_push")
	}
}
