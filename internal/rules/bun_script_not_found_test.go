package rules

import (
	"testing"

	"github.com/leafevans/go-fuck/internal/specific"
)

// bunScriptNotFoundOutput returns the output of bun for a missing script.
func bunScriptNotFoundOutput(script string) string {
	return "error: Script not found \"" + script + "\"\n"
}

// bunStubScripts replaces the package.json scripts with fixture values.
func bunStubScripts(t *testing.T) {
	t.Helper()
	original := specific.PackageScripts
	specific.PackageScripts = func() []string {
		return []string{"build", "deploy", "lint", "test"}
	}
	t.Cleanup(func() { specific.PackageScripts = original })
}

func TestRuleBunScriptNotFound(t *testing.T) {
	resetTestState(t)
	bunStubScripts(t)
	rule := Get("bun_script_not_found")
	if rule == nil {
		t.Fatal("bun_script_not_found rule is not registered")
	}

	assertMatch(t, rule, "bun run buidl", bunScriptNotFoundOutput("buidl"))
	assertMatch(t, rule, "bun buidl", bunScriptNotFoundOutput("buidl"))
	assertNotMatch(t, rule, "npm run buidl", bunScriptNotFoundOutput("buidl"))
	assertNotMatch(t, rule, "bun run build", "error: something else\n")

	got, err := rule.GetNewCommand(testCommand("bun run buidl", bunScriptNotFoundOutput("buidl")))
	if err != nil {
		t.Fatalf("rule %q returned an error: %v", rule.Name, err)
	}
	if len(got) == 0 || got[0] != "bun run build" {
		t.Fatalf("rule %q returned %q, want first command %q", rule.Name, got, "bun run build")
	}
}

func TestRuleBunScriptNotFoundAnsi(t *testing.T) {
	resetTestState(t)
	bunStubScripts(t)
	rule := Get("bun_script_not_found")

	output := "\x1b[31merror\x1b[0m: Script not found \"\x1b[1mbuidl\x1b[0m\"\n"
	assertMatch(t, rule, "bun run buidl", output)

	got, err := rule.GetNewCommand(testCommand("bun run buidl", output))
	if err != nil {
		t.Fatalf("rule %q returned an error: %v", rule.Name, err)
	}
	if len(got) == 0 || got[0] != "bun run build" {
		t.Fatalf("rule %q returned %q, want first command %q", rule.Name, got, "bun run build")
	}
}

func TestRuleBunScriptNotFoundWithoutScripts(t *testing.T) {
	resetTestState(t)
	original := specific.PackageScripts
	specific.PackageScripts = func() []string { return nil }
	t.Cleanup(func() { specific.PackageScripts = original })

	rule := Get("bun_script_not_found")
	assertNotMatch(t, rule, "bun run buidl", bunScriptNotFoundOutput("buidl"))
}
