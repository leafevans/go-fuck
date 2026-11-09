package rules

import "testing"

// rustupMissingToolchainOutput is the error of a toolchain that is not
// installed.
const rustupMissingToolchainOutput = "error: toolchain \x27nightly\x27 is not installed\n"

func TestRuleRustupToolchainMissing(t *testing.T) {
	resetTestState(t)
	rule := Get("rustup_toolchain_missing")
	if rule == nil {
		t.Fatal("rustup_toolchain_missing rule is not registered")
	}

	assertMatch(t, rule, "rustup default nightly", rustupMissingToolchainOutput)
	assertMatch(t, rule, "cargo +nightly build", rustupMissingToolchainOutput)
	assertNotMatch(t, rule, "rustup default nightly", "error: unknown toolchain\n")
	assertNotMatch(t, rule, "rustup default nightly", "")

	assertNewCommands(t, rule, "rustup default nightly", rustupMissingToolchainOutput,
		[]string{"rustup toolchain install nightly && rustup default nightly"})
	assertNewCommands(t, rule, "cargo +nightly build", rustupMissingToolchainOutput,
		[]string{"rustup toolchain install nightly && cargo +nightly build"})
}
