package customrules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/consts"
	"github.com/leafevans/go-fuck/internal/types"
)

// writeRule writes a rule file into dir.
func writeRule(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write rule file: %v", err)
	}
}

func TestLoadRule(t *testing.T) {
	dir := t.TempDir()
	writeRule(t, dir, "git_remote.json", `{
  "name": "git_remote_typo",
  "priority": 900,
  "match": {
    "script": "^git remote add (\\w+)$",
    "output": "No such remote '([^']+)'"
  },
  "commands": ["git remote remove {1}"]
}`)

	rules, problems := Load(dir)
	if len(problems) != 0 {
		t.Fatalf("Load() problems = %v, want none", problems)
	}
	if len(rules) != 1 {
		t.Fatalf("Load() returned %d rules, want 1", len(rules))
	}

	rule := rules[0]
	if rule.Name != "git_remote_typo" {
		t.Fatalf("rule name = %q", rule.Name)
	}
	if rule.Priority != 900 {
		t.Fatalf("rule priority = %d, want 900", rule.Priority)
	}
	if !rule.EnabledByDefault || !rule.RequiresOutput {
		t.Fatalf("rule flags = enabled %t, requires output %t", rule.EnabledByDefault, rule.RequiresOutput)
	}

	command := types.NewCommand("git remote add origin", "fatal: No such remote 'origin'")
	if !rule.IsMatch(command) {
		t.Fatal("rule does not match the command")
	}
	corrected := rule.GetCorrectedCommands(command)
	if len(corrected) != 1 || corrected[0].Script != "git remote remove origin" {
		t.Fatalf("corrected commands = %v", corrected)
	}
}

func TestLoadScriptOnlyRule(t *testing.T) {
	dir := t.TempDir()
	writeRule(t, dir, "echo.json", `{
  "name": "echo_fix",
  "match": {"script": "^echoo (.*)$"},
  "commands": ["echo {script}"]
}`)

	rules, problems := Load(dir)
	if len(problems) != 0 || len(rules) != 1 {
		t.Fatalf("Load() = %d rules, problems %v", len(rules), problems)
	}

	rule := rules[0]
	if rule.RequiresOutput {
		t.Fatal("rule without match.output must not require output")
	}
	command := types.NewCommandWithoutOutput("echoo hello")
	if !rule.IsMatch(command) {
		t.Fatal("rule does not match the command")
	}
	corrected := rule.GetCorrectedCommands(command)
	if len(corrected) != 1 || corrected[0].Script != "echo echoo hello" {
		t.Fatalf("corrected commands = %v", corrected)
	}
}

func TestLoadRuleWithUTF8BOM(t *testing.T) {
	dir := t.TempDir()
	writeRule(t, dir, "bom.json", "\ufeff"+`{"name": "bom_rule", "match": {"script": "^x$"}, "commands": ["y"]}`)

	rules, problems := Load(dir)
	if len(problems) != 0 || len(rules) != 1 {
		t.Fatalf("Load() = %d rules, problems %v", len(rules), problems)
	}
	if rules[0].Name != "bom_rule" {
		t.Fatalf("rule name = %q", rules[0].Name)
	}
}

func TestLoadDisabledRule(t *testing.T) {
	dir := t.TempDir()
	writeRule(t, dir, "disabled.json", `{
  "name": "disabled_rule",
  "enabled": false,
  "match": {"script": "^x$"},
  "commands": ["y"]
}`)

	rules, problems := Load(dir)
	if len(problems) != 0 || len(rules) != 1 {
		t.Fatalf("Load() = %d rules, problems %v", len(rules), problems)
	}
	if rules[0].EnabledByDefault {
		t.Fatal("disabled rule is enabled by default")
	}
	if rules[0].Priority != consts.DefaultPriority {
		t.Fatalf("priority = %d, want the default", rules[0].Priority)
	}
}

func TestLoadProblems(t *testing.T) {
	cases := map[string]string{
		"invalid json":                 `{`,
		"missing name":                 `{"match": {"script": "^x$"}, "commands": ["y"]}`,
		"invalid name":                 `{"name": "Bad Name", "match": {"script": "^x$"}, "commands": ["y"]}`,
		"missing match":                `{"name": "no_match", "commands": ["y"]}`,
		"missing commands":             `{"name": "no_commands", "match": {"script": "^x$"}}`,
		"invalid expression":           `{"name": "bad_regex", "match": {"script": "("}, "commands": ["y"]}`,
		"invalid priority":             `{"name": "bad_priority", "priority": -1, "match": {"script": "^x$"}, "commands": ["y"]}`,
		"placeholder without output":   `{"name": "no_output", "match": {"script": "^x$"}, "commands": ["echo {1}"]}`,
		"placeholder exceeds captures": `{"name": "too_few_groups", "match": {"script": "^x$", "output": "(y)"}, "commands": ["echo {2}"]}`,
		"empty command":                `{"name": "empty_command", "match": {"script": "^x$"}, "commands": [""]}`,
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			writeRule(t, dir, "rule.json", content)

			rules, problems := Load(dir)
			if len(rules) != 0 {
				t.Fatalf("Load() returned %d rules, want none", len(rules))
			}
			if len(problems) != 1 {
				t.Fatalf("Load() problems = %v, want one", problems)
			}
		})
	}
}

func TestLoadDuplicateNames(t *testing.T) {
	dir := t.TempDir()
	writeRule(t, dir, "a.json", `{"name": "duplicate", "match": {"script": "^x$"}, "commands": ["y"]}`)
	writeRule(t, dir, "b.json", `{"name": "duplicate", "match": {"script": "^x$"}, "commands": ["z"]}`)

	rules, problems := Load(dir)
	if len(rules) != 1 {
		t.Fatalf("Load() returned %d rules, want 1", len(rules))
	}
	if len(problems) != 1 || !strings.Contains(problems[0].Error(), "already used") {
		t.Fatalf("Load() problems = %v", problems)
	}
}

func TestLoadMissingDirectory(t *testing.T) {
	rules, problems := Load(filepath.Join(t.TempDir(), "missing"))
	if len(rules) != 0 || len(problems) != 0 {
		t.Fatalf("Load() = %d rules, problems %v", len(rules), problems)
	}
}

func TestLoadIgnoresNonJSONFiles(t *testing.T) {
	dir := t.TempDir()
	writeRule(t, dir, "notes.txt", "not a rule")

	rules, problems := Load(dir)
	if len(rules) != 0 || len(problems) != 0 {
		t.Fatalf("Load() = %d rules, problems %v", len(rules), problems)
	}
}
