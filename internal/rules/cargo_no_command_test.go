package rules

import "testing"

const cargoNoCommandNoSuchSubcommandOld = "No such subcommand\n\n" +
	"        Did you mean `build`?\n"

const cargoNoCommandNoSuchSubcommand = "error: no such subcommand\n\n" +
	"\tDid you mean `build`?\n"

// cargoNoCommandModern is the output of current cargo releases (1.99).
const cargoNoCommandModern = "error: no such command: `buidl`\n\n" +
	"help: a command with a similar name exists: `build`\n\n" +
	"help: view all installed commands with `cargo --list`\n"

// cargoNoCommandUnrecognized is the clap style of some cargo compatible
// tools.
const cargoNoCommandUnrecognized = "error: unrecognized subcommand 'buidl'\n\n" +
	"  tip: a similar subcommand exists: 'build'\n"

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
	assertMatch(t, rule, "cargo buidl", cargoNoCommandModern)
	assertNewCommands(t, rule, "cargo buidl", cargoNoCommandModern,
		[]string{"cargo build"})

	assertMatch(t, rule, "cargo buidl", cargoNoCommandUnrecognized)
	assertNewCommands(t, rule, "cargo buidl", cargoNoCommandUnrecognized,
		[]string{"cargo build"})
}
