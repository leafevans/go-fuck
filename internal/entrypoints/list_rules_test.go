package entrypoints

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/rules"
)

func TestListRules(t *testing.T) {
	resetEntrypointsState(t)

	code, stdout, stderr := runMain(t, "go-fuck", "--list-rules")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0; stderr = %q", code, stderr)
	}
	names := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(names) != rules.Count() {
		t.Fatalf("listed %d rules, want %d", len(names), rules.Count())
	}
	if !sort.StringsAreSorted(names) {
		t.Fatalf("the rule list is not sorted")
	}
	found := false
	for _, name := range names {
		if name == "powershell_remove_recurse" {
			found = true
		}
	}
	if !found {
		t.Fatalf("the rule list does not contain powershell_remove_recurse")
	}
}

func TestListRulesJSON(t *testing.T) {
	resetEntrypointsState(t)

	code, stdout, stderr := runMain(t, "go-fuck", "--list-rules", "--json")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0; stderr = %q", code, stderr)
	}
	var names []string
	if err := json.Unmarshal([]byte(stdout), &names); err != nil {
		t.Fatalf("json.Unmarshal(%q) error = %v", stdout, err)
	}
	if len(names) != rules.Count() {
		t.Fatalf("listed %d rules, want %d", len(names), rules.Count())
	}
}
