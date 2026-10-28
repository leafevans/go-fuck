package rules

import "testing"

// kubectlUnknownCommandOutput is the output of kubectl for a misspelled top
// level command.
func kubectlUnknownCommandOutput() string {
	return `error: unknown command "gt" for "kubectl"

Did you mean this?
	get

Run 'kubectl --help' for usage.
`
}

// kubectlUnknownSubcommandOutput is the output for a misspelled nested
// subcommand.
func kubectlUnknownSubcommandOutput() string {
	return `error: unknown command "vie" for "kubectl config"

Did you mean this?
	view

Run 'kubectl config --help' for usage.
`
}

func TestRuleKubectlUnknownCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("kubectl_unknown_command")
	if rule == nil {
		t.Fatal("kubectl_unknown_command rule is not registered")
	}

	assertMatch(t, rule, "kubectl gt", kubectlUnknownCommandOutput())
	assertMatch(t, rule, "kubectl config vie", kubectlUnknownSubcommandOutput())
	assertNotMatch(t, rule, "helm gt", kubectlUnknownCommandOutput())
	assertNotMatch(t, rule, "kubectl get pods", "No resources found in default namespace.")

	commandCases := []struct {
		script string
		output string
		want   []string
	}{
		{"kubectl gt", kubectlUnknownCommandOutput(), []string{"kubectl get"}},
		{"kubectl config vie", kubectlUnknownSubcommandOutput(), []string{"kubectl config view"}},
	}
	for _, testCase := range commandCases {
		t.Run("new_command/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, testCase.output, testCase.want)
		})
	}
}
