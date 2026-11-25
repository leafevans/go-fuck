package rules

import (
	"strings"
	"testing"
)

// doctlUnknownCommandOutput is the verbatim error that doctl prints for a
// misspelled command; it comes from doctl's integration/errors_test.go.
func doctlUnknownCommandOutput() string {
	return `Error: unknown command "apa" for "doctl"

Did you mean this?
	apps

Run 'doctl --help' for usage.`
}

func TestRuleDoctlUnknownCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("doctl_unknown_command")
	if rule == nil {
		t.Fatal("doctl_unknown_command rule is not registered")
	}

	assertMatch(t, rule, "doctl apa", doctlUnknownCommandOutput())
	assertNotMatch(t, rule, "kubectl apa", doctlUnknownCommandOutput())
	assertNotMatch(t, rule, "doctl account get", "ID  Email  Status")

	assertNewCommands(t, rule, "doctl apa", doctlUnknownCommandOutput(), []string{"doctl apps"})
}

func TestScopeddoctlCorrection(t *testing.T) {
	resetTestState(t)
	rule := Get("doctl_unknown_command")
	output := strings.ReplaceAll(doctlUnknownCommandOutput(), "\n", "\r\n")
	assertNewCommands(t, rule, "doctl apa --context apa", output, []string{"doctl apps --context apa"})
	assertNotMatch(t, rule, "doctl --unknown apa apa", output)
	assertNotMatch(t, rule, "doctl apa --context apa; echo done", output)
}
