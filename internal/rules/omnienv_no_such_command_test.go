package rules

import (
	"testing"

	"github.com/leafevans/go-fuck/internal/utils"
)

// omnienvNoSuchCommandAppCommandsFixture mirrors the output of
// "pyenv commands" used by the original test.
var omnienvNoSuchCommandAppCommandsFixture = []string{
	"--version", "activate", "commands", "completions", "deactivate", "exec_",
	"global", "help", "hooks", "init", "install", "local", "prefix_",
	"realpath.dylib", "rehash", "root", "shell", "shims", "uninstall", "version_",
	"version-file", "version-file-read", "version-file-write", "version-name_",
	"version-origin", "versions", "virtualenv", "virtualenv-delete_",
	"virtualenv-init", "virtualenv-prefix", "virtualenvs_",
	"virtualenvwrapper", "virtualenvwrapper_lazy", "whence", "which_",
}

func TestRuleOmnienvNoSuchCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("omnienv_no_such_command")
	if rule == nil {
		t.Fatal("omnienv_no_such_command rule is not registered")
	}

	originalCommands := omnienvNoSuchCommandAppCommands
	t.Cleanup(func() { omnienvNoSuchCommandAppCommands = originalCommands })
	omnienvNoSuchCommandAppCommands = func(string) []string {
		return append([]string{}, omnienvNoSuchCommandAppCommandsFixture...)
	}

	matchCases := []struct {
		script string
		output string
	}{
		{"pyenv globe", "pyenv: no such command `globe'"},
		{"pyenv intall 3.8.0", "pyenv: no such command `intall'"},
		{"pyenv list", "pyenv: no such command `list'"},
		{"goenv list", "goenv: no such command 'list'"},
	}
	for _, testCase := range matchCases {
		assertMatch(t, rule, testCase.script, testCase.output)
	}

	assertNotMatch(t, rule, "pyenv global", "system")
	assertNotMatch(t, rule, "pyenv versions", "  3.7.0\n  3.7.1\n* 3.7.2\n")
	assertNotMatch(t, rule, "pyenv install --list", "  3.7.0\n  3.7.1\n  3.7.2\n")

	newCommandCases := []struct {
		script string
		output string
		want   string
	}{
		{"pyenv globe", "pyenv: no such command `globe'", "pyenv global"},
		{"pyenv intall 3.8.0", "pyenv: no such command `intall'", "pyenv install 3.8.0"},
		{"pyenv list", "pyenv: no such command `list'", "pyenv install --list"},
		{"pyenv remove 3.8.0", "pyenv: no such command `remove'", "pyenv uninstall 3.8.0"},
	}
	for _, testCase := range newCommandCases {
		got, err := rule.GetNewCommand(testCommand(testCase.script, testCase.output))
		if err != nil {
			t.Fatalf("rule %q returned an error: %v", rule.Name, err)
		}
		if !utils.Contains(got, testCase.want) {
			t.Errorf("rule %q returned %q, want it to contain %q", rule.Name, got, testCase.want)
		}
	}
}
