package rules

import "testing"

// minikubeUnknownCommandOutput is the output of minikube for a misspelled
// command; the suggestions follow the "Did you mean this?" line.
func minikubeUnknownCommandOutput() string {
	return "Error: unknown command \"del\" for \"minikube\"\n" +
		"\n" +
		"Did you mean this?\n" +
		"\tdelete\n" +
		"\n" +
		"Run 'minikube --help' for usage.\n"
}

// minikubeUnknownSubcommandOutput is the output for a misspelled nested
// subcommand.
func minikubeUnknownSubcommandOutput() string {
	return "Error: unknown command \"enabl\" for \"minikube addons\"\n" +
		"\n" +
		"Did you mean this?\n" +
		"\tenable\n" +
		"\n" +
		"Run 'minikube addons --help' for usage.\n"
}

func TestRuleMinikubeUnknownCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("minikube_unknown_command")
	if rule == nil {
		t.Fatal("minikube_unknown_command rule is not registered")
	}

	assertMatch(t, rule, "minikube del", minikubeUnknownCommandOutput())
	assertMatch(t, rule, "minikube addons enabl", minikubeUnknownSubcommandOutput())
	assertNotMatch(t, rule, "kubectl del", minikubeUnknownCommandOutput())
	assertNotMatch(t, rule, "minikube delete", "Deleting \"minikube\" in profile \"minikube\"...\n")
	assertNotMatch(t, rule, "minikube del", "Error: unknown command \"del\" for \"minikube\"\n")

	commandCases := []struct {
		script string
		output string
		want   []string
	}{
		{"minikube del", minikubeUnknownCommandOutput(), []string{"minikube delete"}},
		{"minikube addons enabl", minikubeUnknownSubcommandOutput(), []string{"minikube addons enable"}},
	}
	for _, testCase := range commandCases {
		t.Run("new_command/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, testCase.output, testCase.want)
		})
	}
}
