package rules

import "testing"

// cargoMissingTestTargetOutput is the cargo output for a missing test target.
const cargoMissingTestTargetOutput = "error: no test target named \x60missing\x60 in default-run packages\n" +
	"help: available test targets:\n    foo\n"

// cargoMissingTestTargetMultipleOutput lists several test targets.
const cargoMissingTestTargetMultipleOutput = "error: no test target named \x60helpr\x60 in default-run packages\n" +
	"help: available test targets:\n    foo\n    helper\n"

func TestRuleCargoMissingTestTarget(t *testing.T) {
	resetTestState(t)
	rule := Get("cargo_missing_test_target")
	if rule == nil {
		t.Fatal("cargo_missing_test_target rule is not registered")
	}

	assertMatch(t, rule, "cargo test --test missing", cargoMissingTestTargetOutput)
	assertNotMatch(t, rule, "cargo test --test missing", "")
	assertNotMatch(t, rule, "cargo test --test missing",
		"error: no test target named \x60missing\x60\n")
	assertNotMatch(t, rule, "cargo run --test missing", cargoMissingTestTargetOutput)

	assertNewCommands(t, rule, "cargo test --test missing", cargoMissingTestTargetOutput,
		[]string{"cargo test --test foo"})
	assertNewCommands(t, rule, "cargo test --test helpr", cargoMissingTestTargetMultipleOutput,
		[]string{"cargo test --test helper"})
}
