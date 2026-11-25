package rules

import (
	"strings"
	"testing"
)

// ghUnknownCommandOutput is the output of gh 2.95 for a misspelled top level
// command; the available command list follows the suggestions.
func ghUnknownCommandOutput() string {
	return `unknown command "prr" for "gh"

Did you mean this?
	org
	pr
Usage:  gh <command> <subcommand> [flags]
Available commands:
  agent-task
  alias
  pr
  repo
`
}

// ghUnknownSubcommandOutput is the output for a misspelled nested subcommand.
func ghUnknownSubcommandOutput() string {
	return `unknown command "creat" for "gh repo"

Did you mean this?
	create
Usage:  gh repo <command> [flags]
Available commands:
  archive
  create
  delete
`
}

func TestRuleGhUnknownCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("gh_unknown_command")
	if rule == nil {
		t.Fatal("gh_unknown_command rule is not registered")
	}

	assertMatch(t, rule, "gh prr", ghUnknownCommandOutput())
	assertMatch(t, rule, "gh repo creat", ghUnknownSubcommandOutput())
	assertNotMatch(t, rule, "git prr", ghUnknownCommandOutput())
	assertNotMatch(t, rule, "gh pr", "*+ refs/heads/main")

	commandCases := []struct {
		script string
		output string
		want   []string
	}{
		{"gh prr", ghUnknownCommandOutput(), []string{"gh pr", "gh org"}},
		{"gh repo creat", ghUnknownSubcommandOutput(), []string{"gh repo create"}},
	}
	for _, testCase := range commandCases {
		t.Run("new_command/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, testCase.output, testCase.want)
		})
	}
}

func TestScopedCobraCorrections(t *testing.T) {
	resetTestState(t)
	for _, tc := range []struct{ app, script, output, want string }{
		{"gh", "gh repo creat creat", ghUnknownSubcommandOutput(), "gh repo create creat"},
		{"gh", "gh -R creat repo creat", ghUnknownSubcommandOutput(), ""},
		{"kubectl", "kubectl --context vie config \"vie\" --context vie", kubectlUnknownSubcommandOutput(), "kubectl --context vie config view --context vie"},
		{"kubectl", "kubectl get vie", kubectlUnknownSubcommandOutput(), ""},
		{"helm", "helm --namespace listt repo listt --namespace 'listt'", helmUnknownSubcommandOutput(), "helm --namespace listt repo list --namespace 'listt'"},
		{"helm", "helm repo listt; echo listt", helmUnknownSubcommandOutput(), ""},
	} {
		t.Run(tc.script, func(t *testing.T) {
			rule := Get(tc.app + "_unknown_command")
			output := strings.ReplaceAll(tc.output, "\n", "\r\n")
			if tc.want == "" {
				assertNotMatch(t, rule, tc.script, output)
				assertNewCommands(t, rule, tc.script, output, nil)
				return
			}
			assertMatch(t, rule, tc.script, output)
			assertNewCommands(t, rule, tc.script, output, []string{tc.want})
		})
	}
}
