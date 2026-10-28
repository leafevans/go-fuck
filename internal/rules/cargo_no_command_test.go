package rules

import "testing"

const cargoNoCommandNoSuchSubcommandOld = "No such subcommand\n\n" +
	"        Did you mean `build`?\n"

const cargoNoCommandNoSuchSubcommand = "error: no such subcommand\n\n" +
	"\tDid you mean `build`?\n"

func TestRuleCargoNoCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("cargo_no_command")
	if rule == nil {
		t.Fatal("cargo_no_command rule is not registered")
	}

	assertMatch(t, rule, "cargo buid", cargoNoCommandNoSuchSubcommandOld)
	assertMatch(t, rule, "cargo buils", cargoNoCommandNoSuchSubcommand)

	assertNewCommands(t, rule, "cargo buid", cargoNoCommandNoSuchSubcommandOld,
		[]string{"cargo build"})
	assertNewCommands(t, rule, "cargo buils", cargoNoCommandNoSuchSubcommand,
		[]string{"cargo build"})
}
