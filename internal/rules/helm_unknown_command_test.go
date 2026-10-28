package rules

import "testing"

// helmUnknownCommandOutput is the output of Helm for a misspelled top level
// command.
func helmUnknownCommandOutput() string {
	return "Error: unknown command \"instal\" for \"helm\"\n\n" +
		"Did you mean this?\n" +
		"\tinstall\n" +
		"\n" +
		"Run 'helm --help' for usage.\n"
}

// helmUnknownSubcommandOutput is the output of Helm for a misspelled nested
// subcommand.
func helmUnknownSubcommandOutput() string {
	return "Error: unknown command \"listt\" for \"helm repo\"\n\n" +
		"Did you mean this?\n" +
		"\tlist\n" +
		"\n" +
		"Run 'helm repo --help' for usage.\n"
}

func TestRuleHelmUnknownCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("helm_unknown_command")
	if rule == nil {
		t.Fatal("helm_unknown_command rule is not registered")
	}

	assertMatch(t, rule, "helm instal", helmUnknownCommandOutput())
	assertMatch(t, rule, "helm repo listt", helmUnknownSubcommandOutput())
	assertNotMatch(t, rule, "kubectl instal", helmUnknownCommandOutput())
	assertNotMatch(t, rule, "helm list", "NAME\tNAMESPACE\tREVISION")

	commandCases := []struct {
		script string
		output string
		want   []string
	}{
		{"helm instal", helmUnknownCommandOutput(), []string{"helm install"}},
		{"helm repo listt", helmUnknownSubcommandOutput(), []string{"helm repo list"}},
	}
	for _, testCase := range commandCases {
		t.Run("new_command/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, testCase.output, testCase.want)
		})
	}
}
