package entrypoints

import (
	json "encoding/json/v2"
	"sort"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/rules"
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
	var listed []listRule
	if err := json.Unmarshal([]byte(stdout), &listed); err != nil {
		t.Fatalf("json.Unmarshal(%q) error = %v", stdout, err)
	}
	if len(listed) != rules.Count() {
		t.Fatalf("listed %d rules, want %d", len(listed), rules.Count())
	}
	for index := 1; index < len(listed); index++ {
		if listed[index-1].Name >= listed[index].Name {
			t.Fatalf("the rule list is not sorted at %q", listed[index].Name)
		}
	}
	for _, rule := range listed {
		switch rule.Name {
		case "dry":
			if rule.Priority != 900 || !rule.EnabledByDefault || !rule.RequiresOutput || rule.SideEffect {
				t.Fatalf("unexpected metadata for dry: %+v", rule)
			}
		case "ssh_known_hosts":
			if !rule.SideEffect {
				t.Fatalf("ssh_known_hosts side_effect = false, want true")
			}
		}
	}
}
