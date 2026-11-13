package main

import (
	"bytes"
	"sort"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/rules"
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

func TestPrintRules(t *testing.T) {
	buffer := &bytes.Buffer{}
	printRules(buffer)
	lines := strings.Split(strings.TrimSpace(buffer.String()), "\n")
	if len(lines) != rules.Count() {
		t.Fatalf("printRules() wrote %d lines, want %d", len(lines), rules.Count())
	}
	if !sort.StringsAreSorted(lines) {
		t.Fatal("printRules() output is not sorted")
	}
}
