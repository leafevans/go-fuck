package rules

import "testing"

// cargoMissingBenchTargetOutput is the cargo output for a missing bench target.
const cargoMissingBenchTargetOutput = "error: no bench target named \x60missing\x60 in default-run packages\n" +
	"help: available bench targets:\n    mybench\n"

func TestRuleCargoMissingBenchTarget(t *testing.T) {
	resetTestState(t)
	rule := Get("cargo_missing_bench_target")
	if rule == nil {
		t.Fatal("cargo_missing_bench_target rule is not registered")
	}

	assertMatch(t, rule, "cargo bench --bench missing", cargoMissingBenchTargetOutput)
	assertNotMatch(t, rule, "cargo bench --bench missing", "")
	assertNotMatch(t, rule, "cargo bench --bench missing",
		"error: no bench target named \x60missing\x60\n")
	assertNotMatch(t, rule, "cargo test --bench missing", cargoMissingBenchTargetOutput)

	assertNewCommands(t, rule, "cargo bench --bench missing", cargoMissingBenchTargetOutput,
		[]string{"cargo bench --bench mybench"})
}
