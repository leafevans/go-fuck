package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/utils"
)

func TestRuleWorkonDoesntExists(t *testing.T) {
	resetTestState(t)
	rule := Get("workon_doesnt_exists")
	if rule == nil {
		t.Fatal("workon_doesnt_exists rule is not registered")
	}

	original := workonGetAllEnvironments
	workonGetAllEnvironments = func() []string { return []string{"thefuck", "code_view"} }
	t.Cleanup(func() { workonGetAllEnvironments = original })

	for _, script := range []string{"workon tehfuck", "workon code-view", "workon new-env"} {
		assertMatch(t, rule, script, "")
	}
	for _, script := range []string{"workon thefuck", "workon code_view", "work on tehfuck"} {
		assertNotMatch(t, rule, script, "")
	}

	for _, testCase := range []struct {
		script string
		want   string
	}{
		{"workon tehfuck", "workon thefuck"},
		{"workon code-view", "workon code_view"},
		{"workon zzzz", "mkvirtualenv zzzz"},
	} {
		got, err := rule.GetNewCommand(testCommand(testCase.script, ""))
		if err != nil {
			t.Fatalf("rule %q returned an error: %v", rule.Name, err)
		}
		if len(got) == 0 {
			t.Fatalf("rule %q returned no corrections for %q", rule.Name, testCase.script)
		}
		if got[0] != testCase.want {
			t.Errorf("rule %q returned %q, want %q", rule.Name, got[0], testCase.want)
		}
	}
}

func TestWorkonReadEnvironments(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	envs := filepath.Join(home, ".virtualenvs")
	if err := os.MkdirAll(filepath.Join(envs, "project-one"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(envs, "project-two"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(envs, "not-a-directory"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	got := workonReadEnvironments()
	for _, want := range []string{"project-one", "project-two"} {
		if !utils.Contains(got, want) {
			t.Fatalf("workonReadEnvironments() = %q, want %q", got, want)
		}
	}
	if utils.Contains(got, "not-a-directory") {
		t.Fatalf("workonReadEnvironments() = %q, want directories only", got)
	}
}
