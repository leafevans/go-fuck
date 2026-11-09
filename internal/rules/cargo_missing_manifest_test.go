package rules

import "testing"

// cargoMissingManifestOutput is the output of cargo without a manifest.
const cargoMissingManifestOutput = "error: could not find \x60Cargo.toml\x60 in \x60/tmp/probe\x60 or any parent directory\n"

func TestRuleCargoMissingManifest(t *testing.T) {
	resetTestState(t)
	rule := Get("cargo_missing_manifest")
	if rule == nil {
		t.Fatal("cargo_missing_manifest rule is not registered")
	}

	assertMatch(t, rule, "cargo run", cargoMissingManifestOutput)
	assertMatch(t, rule, "cargo add serde", cargoMissingManifestOutput)
	assertNotMatch(t, rule, "cargo run", "")
	assertNotMatch(t, rule, "rustc --version", cargoMissingManifestOutput)

	assertNewCommands(t, rule, "cargo run", cargoMissingManifestOutput,
		[]string{"cargo init && cargo run"})
}
