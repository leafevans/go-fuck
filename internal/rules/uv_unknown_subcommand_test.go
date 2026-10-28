package rules

import "testing"

// uvUnknownSubcommandOutput is the output of uv for a misspelled top level
// subcommand with a single suggestion.
func uvUnknownSubcommandOutput() string {
	return `error: unrecognized subcommand 'runn'

  tip: a similar subcommand exists: 'run'
Usage: uv.exe [OPTIONS] <COMMAND>
For more information, try '--help'.
`
}

// uvUnknownNestedSubcommandOutput is the output of uv for a misspelled nested
// subcommand with several suggestions.
func uvUnknownNestedSubcommandOutput() string {
	return `error: unrecognized subcommand 'instol'

  tip: some similar subcommands exist: 'uninstall', 'list', 'install'
Usage: uv.exe pip [OPTIONS] <COMMAND>
For more information, try '--help'.
`
}

func TestRuleUvUnknownSubcommand(t *testing.T) {
	resetTestState(t)
	rule := Get("uv_unknown_subcommand")
	if rule == nil {
		t.Fatal("uv_unknown_subcommand rule is not registered")
	}

	assertMatch(t, rule, "uv runn", uvUnknownSubcommandOutput())
	assertMatch(t, rule, "uv pip instol", uvUnknownNestedSubcommandOutput())
	assertNotMatch(t, rule, "pip runn", uvUnknownSubcommandOutput())
	assertNotMatch(t, rule, "uv run main.py", "hello")

	commandCases := []struct {
		script string
		output string
		want   []string
	}{
		{"uv runn", uvUnknownSubcommandOutput(), []string{"uv run"}},
		{"uv pip instol", uvUnknownNestedSubcommandOutput(),
			[]string{"uv pip uninstall", "uv pip list", "uv pip install"}},
	}
	for _, testCase := range commandCases {
		t.Run("new_command/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, testCase.output, testCase.want)
		})
	}
}
