package corrector

import (
	"testing"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/rules"
	"github.com/leafevans/go-fuck/internal/types"
)

// resetCorrectorState resets the settings used by the tests.
func resetCorrectorState(t *testing.T) {
	t.Helper()
	conf.Reset()
	t.Cleanup(conf.Reset)
}

func TestGetRulesOnlyEnabledRules(t *testing.T) {
	resetCorrectorState(t)
	conf.Current().Rules = []string{"quotation_marks"}

	got := []string{}
	for _, rule := range GetRules() {
		got = append(got, rule.Name)
	}
	if len(got) != 1 || got[0] != "quotation_marks" {
		t.Fatalf("GetRules() = %q, want [quotation_marks]", got)
	}
}

func TestGetRulesExcludeRules(t *testing.T) {
	resetCorrectorState(t)
	conf.Current().Rules = []string{"All rules enabled"}
	conf.Current().ExcludeRules = []string{"quotation_marks", "git_push"}

	for _, rule := range GetRules() {
		if rule.Name == "quotation_marks" || rule.Name == "git_push" {
			t.Fatalf("excluded rule %q was returned", rule.Name)
		}
	}
}

func TestGetRulesPriorityOverride(t *testing.T) {
	resetCorrectorState(t)
	conf.Current().Rules = []string{"quotation_marks", "git_push"}
	conf.Current().Priority = map[string]int{"git_push": 1}

	rules := GetRules()
	if len(rules) != 2 {
		t.Fatalf("GetRules() returned %d rules, want 2", len(rules))
	}
	if rules[0].Name != "git_push" || rules[0].Priority != 1 {
		t.Fatalf("unexpected first rule: %+v", rules[0])
	}
}

func TestOrganizeCommands(t *testing.T) {
	commands := []*types.CorrectedCommand{
		{Script: "ls"},
		{Script: "ls -la", Priority: 9000},
		{Script: "ls -lh", Priority: 100},
		{Script: "ls -lh", Priority: 9999},
		{Script: "echo cafe", Priority: 200},
	}

	got := []string{}
	for _, command := range OrganizeCommands(commands) {
		got = append(got, command.Script)
	}
	want := []string{"ls", "ls -lh", "echo cafe", "ls -la"}
	if len(got) != len(want) {
		t.Fatalf("OrganizeCommands() = %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("OrganizeCommands() = %q, want %q", got, want)
		}
	}
}

func TestOrganizeCommandsEmpty(t *testing.T) {
	if got := OrganizeCommands(nil); got != nil {
		t.Fatalf("OrganizeCommands(nil) = %v, want nil", got)
	}
}

func TestGetCorrectedCommands(t *testing.T) {
	resetCorrectorState(t)
	conf.Current().Rules = []string{"quotation_marks"}

	commands := GetCorrectedCommands(types.NewCommand(`git commit -m 'msg"`, ""))
	if len(commands) != 1 {
		t.Fatalf("GetCorrectedCommands() returned %d commands, want 1", len(commands))
	}
	if commands[0].Script != `git commit -m "msg"` {
		t.Fatalf("unexpected command: %q", commands[0].Script)
	}
}

func TestGetCorrectedCommandsNoMatch(t *testing.T) {
	resetCorrectorState(t)
	conf.Current().Rules = []string{"quotation_marks"}

	if commands := GetCorrectedCommands(types.NewCommand("ls", "")); len(commands) != 0 {
		t.Fatalf("GetCorrectedCommands() = %q, want no commands", commands)
	}
}
func TestGetRulesDoesNotMutateRegisteredRules(t *testing.T) {
	resetCorrectorState(t)
	conf.Current().Rules = []string{"quotation_marks", "git_push"}
	conf.Current().Priority = map[string]int{"git_push": 1}
	GetRules()

	// After resetting the priority the registered rule must be untouched.
	conf.Current().Priority = map[string]int{}
	for _, rule := range rules.All() {
		if rule.Name == "git_push" && rule.Priority == 1 {
			t.Fatal("GetRules() mutated the registered rule priority")
		}
	}
}
