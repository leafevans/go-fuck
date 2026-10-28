package rules

import "testing"

const brewReinstallOutput = "Warning: thefuck 9.9 is already installed and up-to-date\n" +
	"To reinstall 9.9, run `brew reinstall thefuck`"

func TestRuleBrewReinstall(t *testing.T) {
	resetTestState(t)
	rule := Get("brew_reinstall")
	if rule == nil {
		t.Fatal("brew_reinstall rule is not registered")
	}

	assertMatch(t, rule, "brew install thefuck", brewReinstallOutput)
	assertNotMatch(t, rule, "brew reinstall thefuck", "")
	assertNotMatch(t, rule, "brew install foo", "")
	assertNotMatch(t, rule, "brew install thefuck", "")

	assertNewCommands(t, rule, "brew install foo", brewReinstallOutput,
		[]string{"brew reinstall foo"})
	assertNewCommands(t, rule, "brew install bar zap", brewReinstallOutput,
		[]string{"brew reinstall bar zap"})
}
