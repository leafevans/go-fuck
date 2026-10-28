package rules

import "testing"

// cargoUnknownArgumentOutput is the output of cargo for a misspelled argument.
func cargoUnknownArgumentOutput() string {
	return "error: unexpected argument '--relase' found\n\n" +
		"  tip: a similar argument exists: '--release'\n" +
		"Usage: cargo build --release\n" +
		"For more information, try '--help'.\n"
}

func TestRuleCargoUnknownArgument(t *testing.T) {
	resetTestState(t)
	rule := Get("cargo_unknown_argument")
	if rule == nil {
		t.Fatal("cargo_unknown_argument rule is not registered")
	}

	assertMatch(t, rule, "cargo build --relase", cargoUnknownArgumentOutput())
	assertNotMatch(t, rule, "rustc build --relase", cargoUnknownArgumentOutput())
	assertNotMatch(t, rule, "cargo build --release", "Compiling demo v0.1.0")

	commandCases := []struct {
		script string
		want   []string
	}{
		{"cargo build --relase", []string{"cargo build --release"}},
		{"cargo build --relase -v", []string{"cargo build --release -v"}},
	}
	for _, testCase := range commandCases {
		t.Run("new_command/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, cargoUnknownArgumentOutput(), testCase.want)
		})
	}
}
