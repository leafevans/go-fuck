package rules

import "testing"

// rustupComponentUnknownOutput is the output of rustup for a misspelled
// top level component subcommand with a single suggestion.
func rustupComponentUnknownOutput() string {
	return `error: unrecognized subcommand 'ad'

  tip: a similar subcommand exists: 'add'

Usage: rustup[EXE] component <COMMAND>

For more information, try '--help'.
`
}

// rustupToolchainUnknownOutput is the output of rustup for a misspelled
// toolchain subcommand with several suggestions.
func rustupToolchainUnknownOutput() string {
	return `error: unrecognized subcommand 'uninstal'

  tip: some similar subcommands exist: 'list', 'install', 'uninstall'

Usage: rustup[EXE] toolchain <COMMAND>

For more information, try '--help'.
`
}

// rustupShowOutput is the output of a successful rustup show command.
func rustupShowOutput() string {
	return `Default host: x86_64-pc-windows-msvc
rustup home:  C:\Users\example\.rustup

installed toolchains
--------------------
stable-x86_64-pc-windows-msvc (active, default)
`
}

func TestRuleRustupUnknownSubcommand(t *testing.T) {
	resetTestState(t)
	rule := Get("rustup_unknown_subcommand")
	if rule == nil {
		t.Fatal("rustup_unknown_subcommand rule is not registered")
	}

	assertMatch(t, rule, "rustup component ad", rustupComponentUnknownOutput())
	assertMatch(t, rule, "rustup toolchain uninstal", rustupToolchainUnknownOutput())
	assertNotMatch(t, rule, "rustup show", rustupShowOutput())
	assertNotMatch(t, rule, "rustup component", "Modify a toolchain's installed components")
	assertNotMatch(t, rule, "uv runn", rustupComponentUnknownOutput())
	assertNotMatch(t, rule, "rustup updat", "error: invalid value 'updat' for '[+toolchain]'")

	commandCases := []struct {
		script string
		output string
		want   []string
	}{
		{"rustup component ad", rustupComponentUnknownOutput(), []string{"rustup component add"}},
		{"rustup +stable component ad", rustupComponentUnknownOutput(), []string{"rustup +stable component add"}},
		{"rustup toolchain uninstal", rustupToolchainUnknownOutput(),
			[]string{"rustup toolchain uninstall", "rustup toolchain install", "rustup toolchain list"}},
	}
	for _, testCase := range commandCases {
		t.Run("new_command/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, testCase.output, testCase.want)
		})
	}
}

// rustupInferredSubcommandOutput is the output of rustup when the misspelled
// prefix of a subcommand is accepted and the next argument is reported as the
// unknown subcommand instead.
func rustupInferredSubcommandOutput() string {
	return `error: unrecognized subcommand 'stable'

  tip: a similar subcommand exists: 'install'

Usage: rustup[EXE] [OPTIONS] [+toolchain] [COMMAND]

For more information, try '--help'.
`
}

// TestRuleRustupUnknownSubcommandInferred checks that the rule does not replace
// the argument that a parser reported after accepting a misspelled prefix.
func TestRuleRustupUnknownSubcommandInferred(t *testing.T) {
	resetTestState(t)
	rule := Get("rustup_unknown_subcommand")
	if rule == nil {
		t.Fatal("rustup_unknown_subcommand rule is not registered")
	}

	output := rustupInferredSubcommandOutput()
	if got, err := rule.GetNewCommand(testCommand("rustup instal stable", output)); err != nil || len(got) != 0 {
		t.Fatalf("GetNewCommand() = %q, %v, want no suggestion", got, err)
	}
}
