package rules

import (
	"testing"

	"github.com/leafevans/go-fuck/internal/types"
)

// noCommandFakeExecutables returns the installed executables used by the tests.
func noCommandFakeExecutables() []string {
	return []string{"vim", "fsck", "git", "go", "python"}
}

// noCommandFakeHistory returns a fixed history for the no_command tests.
func noCommandFakeHistory(*types.Command) []string {
	return []string{"git commit"}
}

// noCommandSetup installs the fake dependencies and restores them afterwards.
func noCommandSetup(t *testing.T) {
	t.Helper()
	originalWhich := noCommandWhich
	originalExecutables := noCommandGetAllExecutables
	originalHistory := noCommandGetValidHistory
	noCommandWhich = func(string) string { return "" }
	noCommandGetAllExecutables = noCommandFakeExecutables
	noCommandGetValidHistory = noCommandFakeHistory
	t.Cleanup(func() {
		noCommandWhich = originalWhich
		noCommandGetAllExecutables = originalExecutables
		noCommandGetValidHistory = originalHistory
	})
}

func TestRuleNoCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("no_command")
	if rule == nil {
		t.Fatal("no_command rule is not registered")
	}
	noCommandSetup(t)

	for _, testCase := range []struct {
		script string
		output string
	}{
		{"vom file.py", "vom: not found"},
		{"fucck", "fucck: not found"},
		{"puthon", "'puthon' is not recognized as an internal or external command"},
		{"got commit", "got: command not found"},
		{"gti commit -m \"new commit\"", "gti: command not found"},
	} {
		assertMatch(t, rule, testCase.script, testCase.output)
	}

	for _, testCase := range []struct {
		script string
		output string
		which  string
	}{
		{"qweqwe", "qweqwe: not found", ""},
		{"vom file.py", "some text", ""},
		{"vim file.py", "vim: not found", "/usr/bin/vim"},
	} {
		noCommandWhich = func(string) string { return "" }
		if testCase.which != "" {
			noCommandWhich = func(name string) string {
				if name == "vim" {
					return testCase.which
				}
				return ""
			}
		}
		assertNotMatch(t, rule, testCase.script, testCase.output)
	}
}

func TestRuleNoCommandNewCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("no_command")
	if rule == nil {
		t.Fatal("no_command rule is not registered")
	}
	noCommandSetup(t)

	for _, testCase := range []struct {
		script string
		want   []string
	}{
		{"vom file.py", []string{"vim file.py"}},
		{"fucck", []string{"fsck"}},
		{"got commit", []string{"git commit", "go commit"}},
		{"gti commit -m \"new commit\"", []string{"git commit -m \"new commit\""}},
	} {
		assertNewCommands(t, rule, testCase.script, "", testCase.want)
	}
}
