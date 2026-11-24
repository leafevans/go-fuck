package entrypoints

import (
	"bytes"
	json "encoding/json/v2"
	"sort"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v3/internal/rules"
	"github.com/leafevans/go-fuck/v3/internal/types"
)

func TestPrintTextRulesTerminalLayout(t *testing.T) {
	all := []*types.Rule{{Name: "git_push"}, {Name: "git_status"}}

	var terminal bytes.Buffer
	printTextRules(&terminal, all, true)
	for _, want := range []string{"go-fuck rules", "INFO", "2 built-in rules", "  git_push", "  git_status"} {
		if !strings.Contains(terminal.String(), want) {
			t.Fatalf("terminal output %q does not contain %q", terminal.String(), want)
		}
	}

	var piped bytes.Buffer
	printTextRules(&piped, all, false)
	if got, want := piped.String(), "git_push\ngit_status\n"; got != want {
		t.Fatalf("piped output = %q, want %q", got, want)
	}
}

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
		case "git_commit_reset", "git_commit_amend", "port_already_in_use":
			if rule.EnabledByDefault {
				t.Fatalf("unsafe legacy default: %+v", rule)
			}
		case "git_hook_bypass":
			if !rule.RequiresOutput {
				t.Fatal("hook bypass must require diagnostics")
			}
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

// TestListRulesKeepsRegistryOrder guards that listing the rules does not sort the
// shared slice returned by rules.All in place.
func TestListRulesKeepsRegistryOrder(t *testing.T) {
	resetEntrypointsState(t)
	before := append([]*types.Rule(nil), rules.All()...)
	code, _, stderr := runMain(t, "go-fuck", "--list-rules")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0; stderr = %q", code, stderr)
	}
	after := rules.All()
	for index := range before {
		if before[index] != after[index] {
			t.Fatalf("--list-rules reordered the shared rule registry at index %d", index)
		}
	}
}
