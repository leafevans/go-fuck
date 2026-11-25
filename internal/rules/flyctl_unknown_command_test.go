package rules

import "testing"

// flyctlUnknownCommandOutput follows the Cobra error that flyctl's
// internal/command/root tests assert for a misspelled subcommand of a group.
func flyctlUnknownCommandOutput() string {
	return `Error: unknown command "lst" for "fly mpg"

Did you mean this?
	list

Run 'fly mpg --help' for usage.`
}

func TestRuleFlyctlUnknownCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("flyctl_unknown_command")
	if rule == nil {
		t.Fatal("flyctl_unknown_command rule is not registered")
	}

	assertMatch(t, rule, "flyctl mpg lst", flyctlUnknownCommandOutput())
	assertMatch(t, rule, "fly mpg lst", flyctlUnknownCommandOutput())
	assertNotMatch(t, rule, "kubectl mpg lst", flyctlUnknownCommandOutput())
	assertNotMatch(t, rule, "flyctl status", "App is running")

	assertNewCommands(t, rule, "flyctl mpg lst", flyctlUnknownCommandOutput(), []string{"flyctl mpg list"})
}
