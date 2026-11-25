package rules

import "testing"

// pdmUnknownCommandOutput follows src/pdm/cli/utils.py (format_similar_command)
// and the argparse error() call in src/pdm/core.py for an unknown subcommand.
func pdmUnknownCommandOutput() string {
	return `usage: pdm [-h] [-V] [--no-cache] ...

pdm: error: Command not found: instal
  Did you mean one of these commands?
  - install
  - export
`
}

func TestRulePdmUnknownCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("pdm_unknown_command")
	if rule == nil {
		t.Fatal("pdm_unknown_command rule is not registered")
	}

	assertMatch(t, rule, "pdm instal", pdmUnknownCommandOutput())
	assertMatch(t, rule, "pdm instal --dev", pdmUnknownCommandOutput())
	assertNotMatch(t, rule, "pip instal", pdmUnknownCommandOutput())
	assertNotMatch(t, rule, "pdm install", "All packages are synced")

	assertNewCommands(t, rule, "pdm instal", pdmUnknownCommandOutput(), []string{"pdm install", "pdm export"})
	assertNewCommands(t, rule, "pdm instal --dev", pdmUnknownCommandOutput(), []string{"pdm install --dev", "pdm export --dev"})
}

func TestRulePdmUnknownCommandANSIAndScripts(t *testing.T) {
	resetTestState(t)
	rule := Get("pdm_unknown_command")

	colored := "pdm: error: \x1b[31mCommand not found: instal\x1b[0m\n" +
		"  \x1b[32mDid you mean one of these commands?\n  - install\x1b[0m\n"
	assertMatch(t, rule, "pdm instal", colored)
	assertNewCommands(t, rule, "pdm instal", colored, []string{"pdm install"})

	withScripts := pdmUnknownCommandOutput() +
		"  Or one of these script commands?\n  - install-all\n"
	assertMatch(t, rule, "pdm instal", withScripts)
	assertNewCommands(t, rule, "pdm instal", withScripts, []string{"pdm install", "pdm export"})
}
