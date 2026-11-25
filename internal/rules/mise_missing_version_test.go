package rules

import (
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v4/internal/shells"
)

// Source-derived fixture: jdx/mise v2025.12.0, src/shims.rs:392-401
// (shim diagnostic), src/main.rs (display_friendly_err) and src/logger.rs
// (default prefix). The node version instantiates the upstream format; this is
// not a captured invocation and does not require mise or network access.
const miseMissingVersionFixture = `mise ERROR Tool not installed for shim: node
Missing tool version: core:node@22.14.0
Install all missing tools with: mise install
mise ERROR Run with --verbose or MISE_VERBOSE=1 for more information
`

func TestRuleMiseMissingVersion(t *testing.T) {
	resetTestState(t)
	rule := Get("mise_missing_version")
	if rule.EnabledByDefault || !rule.RequiresOutput || rule.SideEffect != nil {
		t.Fatal("expected an opt-in, output-dependent rule with no side effect")
	}

	const spec = "core:node@22.14.0"
	plural := strings.Replace(miseMissingVersionFixture, "Tool not installed", "Tools not installed", 1)
	plural = strings.Replace(plural, spec+"\n", spec+"\nMissing tool version: core:node@24.1.0\n", 1)
	for _, tc := range []struct {
		name, script, output, install string
	}{
		{"single", `node --eval "process.version"`, miseMissingVersionFixture, "mise install " + spec},
		{"multiple", "node --version", plural, "mise install " + spec + " core:node@24.1.0"},
		{"color and CRLF", "node -v", strings.ReplaceAll(strings.ReplaceAll(miseMissingVersionFixture, "mise ERROR", "\x1b[31mmise\x1b[0m \x1b[31mERROR\x1b[0m"), "\n", "\r\n"), "mise install " + spec},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertMatch(t, rule, tc.script, tc.output)
			assertNewCommands(t, rule, tc.script, tc.output, []string{tc.install + " && " + tc.script})
		})
	}

	t.Run("PowerShell", func(t *testing.T) {
		shells.Set(shells.NewPowershell())
		defer shells.Set(shells.NewGeneric())
		assertNewCommands(t, rule, "node -v", miseMissingVersionFixture,
			[]string{"mise install " + spec + "; if ($?) { node -v }"})
	})

	for _, tc := range []struct{ name, script, output string }{
		{"unrelated", "node -v", "node: command not found"},
		{"shim mismatch", "python -V", miseMissingVersionFixture},
		{"wrapper", "sudo node -v", miseMissingVersionFixture},
		{"compound", "node -v; echo done", miseMissingVersionFixture},
		{"trust failure", "node -v", "mise ERROR config file ./mise.toml is not trusted.\n" + miseMissingVersionFixture},
		{"no configured version", "node -v", "mise ERROR No version is set for shim: node\nSet a global default version with one of the following:\nmise use -g node@22.14.0"},
		{"truncated", "node -v", strings.Split(miseMissingVersionFixture, "Install all missing")[0]},
		{"missing version line", "node -v", strings.Replace(miseMissingVersionFixture, "Missing tool version: "+spec+"\n", "", 1)},
		{"malformed plural", "node -v", strings.Replace(miseMissingVersionFixture, "Tool not installed", "Tools not installed", 1)},
		{"moving version", "node -v", strings.Replace(miseMissingVersionFixture, spec, "node@latest", 1)},
		{"unsafe tool", "node -v", strings.Replace(miseMissingVersionFixture, spec, "node@22.14.0; echo injected", 1)},
		{"flag as tool", "node -v", strings.Replace(miseMissingVersionFixture, spec, "--force@22.14.0", 1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertNotMatch(t, rule, tc.script, tc.output)
			assertNewCommands(t, rule, tc.script, tc.output, nil)
		})
	}
}
