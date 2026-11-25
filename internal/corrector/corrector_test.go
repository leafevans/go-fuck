package corrector

import (
	"slices"
	"testing"

	"github.com/leafevans/go-fuck/v4/internal/conf"
	"github.com/leafevans/go-fuck/v4/internal/consts"
	"github.com/leafevans/go-fuck/v4/internal/rules"
	"github.com/leafevans/go-fuck/v4/internal/types"
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

func TestGetRulesDefaultAndOptIn(t *testing.T) {
	resetCorrectorState(t)
	settings := conf.Current()
	settings.Rules = []string{consts.AllEnabled}
	want := slices.DeleteFunc(slices.Clone(rules.All()), func(rule *types.Rule) bool {
		return !rule.DefaultEnabled()
	})
	if got := GetRules(); !slices.Equal(got, want) {
		t.Fatal("default rule selection changed membership or order")
	}

	optIn := rules.Get("rm_root")
	if optIn == nil || optIn.DefaultEnabled() {
		t.Fatal("expected rm_root to require opt-in")
	}
	settings.Rules = append(settings.Rules, optIn.Name)
	if !slices.Contains(GetRules(), optIn) {
		t.Fatal("explicitly enabled rule is missing")
	}
	settings.ExcludeRules = []string{optIn.Name}
	if got := GetRules(); !slices.Equal(got, want) {
		t.Fatal("exclusion must take precedence over explicit opt-in")
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

func organizeEffectOne(*types.Command, string) error { return nil }
func organizeEffectTwo(*types.Command, string) error { return nil }

func TestOrganizeCommandsPreservesSideEffectsAndInput(t *testing.T) {
	resetCorrectorState(t)
	commands := []*types.CorrectedCommand{
		{Script: "git status", Priority: 100},
		{Script: "git status", SideEffect: organizeEffectOne, Priority: 20},
		{Script: "git status", SideEffect: organizeEffectTwo, Priority: 10},
		{Script: "git status", SideEffect: organizeEffectOne, Priority: 5},
		{Script: "git status", Priority: 1},
		{Script: "git diff", Priority: 5},
	}
	original := append([]*types.CorrectedCommand(nil), commands...)
	got := OrganizeCommands(commands)
	want := []*types.CorrectedCommand{commands[0], commands[3], commands[5], commands[2]}
	if len(got) != len(want) {
		t.Fatalf("OrganizeCommands() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("suggestion %d = %v, want %v", i, got[i], want[i])
		}
	}
	for i := range commands {
		if commands[i] != original[i] {
			t.Fatalf("input order changed at %d", i)
		}
	}
}

// TestGetCorrectedCommandsRecordsRuleName checks that every suggestion carries
// the name of the rule that produced it, which the machine readable output
// reports.
func TestGetCorrectedCommandsRecordsRuleName(t *testing.T) {
	resetCorrectorState(t)
	conf.Current().Rules = []string{"git_not_command"}

	output := "git: 'brnch' is not a git command. See 'git --help'.\n\n" +
		"The most similar command is\n\tbranch\n"
	command := types.NewCommand("git brnch", output)
	corrected := GetCorrectedCommands(command)
	if len(corrected) == 0 {
		t.Fatal("GetCorrectedCommands() returned no suggestions")
	}
	for _, suggestion := range corrected {
		if suggestion.RuleName != "git_not_command" {
			t.Fatalf("rule name = %q, want git_not_command", suggestion.RuleName)
		}
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

func TestGetRulesKeepsTieOrderWithPriorityOverrides(t *testing.T) {
	resetCorrectorState(t)
	conf.Current().Rules = []string{"quotation_marks", "git_push"}
	registered := rules.Get("git_push")
	priority := registered.Priority
	conf.Current().Priority = map[string]int{"git_push": priority, "quotation_marks": priority}
	var want []string
	for _, rule := range rules.All() {
		if rule.Name == "git_push" || rule.Name == "quotation_marks" {
			want = append(want, rule.Name)
		}
	}
	got := GetRules()
	if len(got) != 2 || got[0].Name != want[0] || got[1].Name != want[1] || got[0].Priority != got[1].Priority {
		t.Fatalf("rules with tied priorities = %v, want %v at priority %d", got, want, priority)
	}
	for _, rule := range got {
		if rule == registered {
			t.Fatal("overridden rule must not share the registered instance")
		}
	}
}

func TestAppendCorrections(t *testing.T) {
	command := types.NewCommand("git status", "")
	sideEffect := func(*types.Command, string) error { return nil }
	candidates := []*types.CorrectedCommand{
		{Script: "git status"},
		{Script: "git status", SideEffect: sideEffect},
		{Script: "git status --short"},
	}

	prior := &types.CorrectedCommand{Script: "git diff", RuleName: "previous"}
	got := appendCorrections([]*types.CorrectedCommand{prior}, command.Script, "example", candidates)
	if !slices.Equal(got, []*types.CorrectedCommand{prior, candidates[1], candidates[2]}) {
		t.Fatalf("appendCorrections() = %q, want previous and both useful suggestions", got)
	}
	if prior.RuleName != "previous" || candidates[0].RuleName != "" || got[1].RuleName != "example" || got[2].RuleName != "example" {
		t.Fatalf("rule name assigned incorrectly: %q, dropped candidate: %q", got, candidates[0])
	}
}

func TestOrganizeCommandsDeduplicationPathsAgree(t *testing.T) {
	resetCorrectorState(t)
	first := &types.CorrectedCommand{Script: "same", Priority: 99}
	withEffect := &types.CorrectedCommand{Script: "same", SideEffect: organizeEffectOne, Priority: 2}
	tied := &types.CorrectedCommand{Script: "other", Priority: 2}
	commands := []*types.CorrectedCommand{first, withEffect, tied, {Script: "same", Priority: 1}, {Script: "same", SideEffect: organizeEffectOne, Priority: 3}}
	want := []*types.CorrectedCommand{first, withEffect, tied}
	for _, size := range []int{5, 16, 17, 32} {
		input := slices.Clone(commands)
		for len(input) < size {
			input = append(input, commands[4])
		}
		if got := OrganizeCommands(input); !slices.Equal(got, want) {
			t.Errorf("OrganizeCommands() with %d inputs = %v, want %v", size, got, want)
		}
	}
}

func TestGetCorrectedCommandsAppliesPriorityOverride(t *testing.T) {
	resetCorrectorState(t)
	conf.Current().Rules = []string{"git_push"}
	conf.Current().Priority = map[string]int{"git_push": 7}
	got := GetCorrectedCommands(types.NewCommand("git push", benchmarkGitPushOutput()))
	if len(got) != 1 || got[0].Priority != 7 || got[0].RuleName != "git_push" {
		t.Fatalf("GetCorrectedCommands() = %v, want git_push correction with priority 7", got)
	}
}
