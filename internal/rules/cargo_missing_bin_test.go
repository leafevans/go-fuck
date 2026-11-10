package rules

import "testing"

// cargoMissingBinSingleOutput is the cargo output for a missing bin target in
// a package with one binary.
const cargoMissingBinSingleOutput = "error: no bin target named \x60missing\x60 in default-run packages\n" +
	"help: available bin targets:\n    app\n"

// cargoMissingBinMultipleOutput is the cargo output for a missing bin target
// in a package with several binaries.
const cargoMissingBinMultipleOutput = "error: no bin target named \x60helpr\x60 in default-run packages\n" +
	"help: available bin targets:\n    app\n    helper\n"

func TestRuleCargoMissingBin(t *testing.T) {
	resetTestState(t)
	rule := Get("cargo_missing_bin")
	if rule == nil {
		t.Fatal("cargo_missing_bin rule is not registered")
	}

	assertMatch(t, rule, "cargo run --bin missing", cargoMissingBinSingleOutput)
	assertNotMatch(t, rule, "cargo run --bin missing", "")
	assertNotMatch(t, rule, "cargo run --bin missing", "error: no bin target named \x60missing\x60\n")
	assertNotMatch(t, rule, "rustc missing", cargoMissingBinSingleOutput)

	assertNewCommands(t, rule, "cargo run --bin missing", cargoMissingBinSingleOutput,
		[]string{"cargo run --bin app"})
	assertNewCommands(t, rule, "cargo run --bin helpr", cargoMissingBinMultipleOutput,
		[]string{"cargo run --bin helper"})
}
