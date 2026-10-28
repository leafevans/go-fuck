package rules

import "testing"

// pipUnknownCommandOutput returns the output pip prints for a misspelled
// command.
func pipUnknownCommandOutput(broken, suggested string) string {
	return "ERROR: unknown command \"" + broken + "\" - maybe you meant \"" + suggested + "\""
}

func TestRulePipUnknownCommand(t *testing.T) {
	resetTestState(t)

	rule := Get("pip_unknown_command")
	if rule == nil {
		t.Fatal("pip_unknown_command rule is not registered")
	}

	assertMatch(t, rule, "pip instatl", pipUnknownCommandOutput("instatl", "install"))
	assertNotMatch(t, rule, "pip i", `ERROR: unknown command "i"`)

	assertNewCommands(t, rule, "pip un+install thefuck",
		pipUnknownCommandOutput("un+install", "uninstall"),
		[]string{"pip uninstall thefuck"})
	assertNewCommands(t, rule, "pip instatl",
		pipUnknownCommandOutput("instatl", "install"),
		[]string{"pip install"})

	// pip2 and pip3 are supported as well.
	assertMatch(t, rule, "pip3 instatl", pipUnknownCommandOutput("instatl", "install"))
}
