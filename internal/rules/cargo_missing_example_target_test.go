package rules

import "testing"

// cargoMissingExampleTargetOutput is the cargo output for a missing example
// target.
const cargoMissingExampleTargetOutput = "error: no example target named \x60missing\x60 in default-run packages\n" +
	"help: available example targets:\n    myexample\n"

func TestRuleCargoMissingExampleTarget(t *testing.T) {
	resetTestState(t)
	rule := Get("cargo_missing_example_target")
	if rule == nil {
		t.Fatal("cargo_missing_example_target rule is not registered")
	}

	assertMatch(t, rule, "cargo build --example missing", cargoMissingExampleTargetOutput)
	assertMatch(t, rule, "cargo run --example missing", cargoMissingExampleTargetOutput)
	assertNotMatch(t, rule, "cargo run --example missing", "")
	assertNotMatch(t, rule, "cargo run --example missing",
		"error: no example target named \x60missing\x60\n")

	assertNewCommands(t, rule, "cargo build --example missing", cargoMissingExampleTargetOutput,
		[]string{"cargo build --example myexample"})
	assertNewCommands(t, rule, "cargo run --example missing", cargoMissingExampleTargetOutput,
		[]string{"cargo run --example myexample"})
}
