package rules

import "testing"

// cargoUnknownCommandOutput is the output of cargo for a misspelled
// subcommand.
func cargoUnknownCommandOutput() string {
	return "error: no such command: \x60instll\x60\n\n" +
		"help: a command with a similar name exists: \x60install\x60\n" +
		"help: view all installed commands with \x60cargo --list\x60\n" +
		"help: find a package to install \x60instll\x60 with \x60cargo search cargo-instll\x60\n"
}

func TestRuleCargoUnknownCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("cargo_unknown_command")
	if rule == nil {
		t.Fatal("cargo_unknown_command rule is not registered")
	}

	assertMatch(t, rule, "cargo instll", cargoUnknownCommandOutput())
	assertNotMatch(t, rule, "rustc instll", cargoUnknownCommandOutput())
	assertNotMatch(t, rule, "cargo build", "error: could not compile `demo`")

	assertNewCommands(t, rule, "cargo instll", cargoUnknownCommandOutput(), []string{"cargo install"})
}
