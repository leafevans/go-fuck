package rules

import (
	"fmt"
	"testing"
)

func TestRuleSudoCommandFromUserPath(t *testing.T) {
	resetTestState(t)
	rule := Get("sudo_command_from_user_path")
	if rule == nil {
		t.Fatal("sudo_command_from_user_path rule is not registered")
	}

	original := sudoCommandFromUserPathWhich
	sudoCommandFromUserPathWhich = func(string) string { return "/usr/bin/app" }
	t.Cleanup(func() { sudoCommandFromUserPathWhich = original })

	output := func(name string) string {
		return fmt.Sprintf("sudo: %s: command not found", name)
	}

	for _, testCase := range []struct {
		script string
		output string
	}{
		{"sudo npm install -g react-native-cli", output("npm")},
		{"sudo -u app appcfg update .", output("appcfg")},
	} {
		assertMatch(t, rule, testCase.script, testCase.output)
	}

	for _, testCase := range []struct {
		script string
		output string
		which  string
	}{
		{"npm --version", output("npm"), "/usr/bin/npm"},
		{"sudo npm --version", "", "/usr/bin/npm"},
		{"sudo npm --version", output("npm"), ""},
	} {
		sudoCommandFromUserPathWhich = func(string) string { return testCase.which }
		assertNotMatch(t, rule, testCase.script, testCase.output)
	}
	sudoCommandFromUserPathWhich = func(string) string { return "/usr/bin/app" }

	for _, testCase := range []struct {
		script string
		output string
		want   string
	}{
		{"sudo npm install -g react-native-cli", output("npm"),
			`sudo env "PATH=$PATH" npm install -g react-native-cli`},
		{"sudo -u app appcfg update .", output("appcfg"),
			`sudo -u app env "PATH=$PATH" appcfg update .`},
	} {
		assertNewCommands(t, rule, testCase.script, testCase.output, []string{testCase.want})
	}
}
