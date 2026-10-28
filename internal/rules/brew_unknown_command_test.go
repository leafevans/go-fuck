package rules

import (
	"testing"

	"github.com/leafevans/go-fuck/internal/utils"
)

const brewUnknownCommandUnknown = `Error: Unknown command: inst`
const brewUnknownCommandUnknownTwo = `Error: Unknown command: instaa`

func TestRuleBrewUnknownCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("brew_unknown_command")
	if rule == nil {
		t.Fatal("brew_unknown_command rule is not registered")
	}

	originalCommands := brewUnknownCommandBrewCommands
	t.Cleanup(func() { brewUnknownCommandBrewCommands = originalCommands })
	brewUnknownCommandBrewCommands = func() []string {
		return append([]string{}, brewUnknownCommandFallbackCommands...)
	}

	assertMatch(t, rule, "brew inst", brewUnknownCommandUnknown)
	for _, command := range brewUnknownCommandBrewCommands() {
		assertNotMatch(t, rule, "brew "+command, "")
	}

	assertNewCommands(t, rule, "brew inst", brewUnknownCommandUnknown,
		[]string{"brew list", "brew install", "brew uninstall"})

	commands, err := rule.GetNewCommand(testCommand("brew instaa", brewUnknownCommandUnknownTwo))
	if err != nil {
		t.Fatalf("rule %q returned an error: %v", rule.Name, err)
	}
	for _, want := range []string{"brew install", "brew uninstall"} {
		if !utils.Contains(commands, want) {
			t.Errorf("rule %q returned %q, want it to contain %q", rule.Name, commands, want)
		}
	}
}
