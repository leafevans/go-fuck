package rules

import "testing"

// uvUnknownArgumentOutput is the output of uv for an unknown argument.
func uvUnknownArgumentOutput() string {
	return "error: unexpected argument '--versio' found\n" +
		"\n" +
		"  tip: a similar argument exists: '--version'\n" +
		"Usage: uv.exe --version <COMMAND>\n" +
		"For more information, try '--help'.\n"
}

func TestRuleUvUnknownArgument(t *testing.T) {
	resetTestState(t)
	rule := Get("uv_unknown_argument")
	if rule == nil {
		t.Fatal("uv_unknown_argument rule is not registered")
	}

	assertMatch(t, rule, "uv --versio", uvUnknownArgumentOutput())
	assertNotMatch(t, rule, "pip --versio", uvUnknownArgumentOutput())
	assertNotMatch(t, rule, "uv --version", "uv 0.9.0")

	assertNewCommands(t, rule, "uv --versio", uvUnknownArgumentOutput(), []string{"uv --version"})
}
