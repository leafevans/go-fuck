package rules

import "testing"

// pnpmMissingScriptOutput is the output of pnpm for a misspelled package
// script.
func pnpmMissingScriptOutput() string {
	return "undefined\n" +
		" ERR_PNPM_RECURSIVE_EXEC_FIRST_FAIL  Command \"buld\" not found\n" +
		"\n" +
		"Did you mean \"pnpm build\"?\n"
}

func TestRulePnpmCommandNotFound(t *testing.T) {
	resetTestState(t)
	rule := Get("pnpm_command_not_found")
	if rule == nil {
		t.Fatal("pnpm_command_not_found rule is not registered")
	}

	assertMatch(t, rule, "pnpm buld", pnpmMissingScriptOutput())
	assertNotMatch(t, rule, "npm buld", pnpmMissingScriptOutput())
	assertNotMatch(t, rule, "pnpm build", "> fixture@1.0.0 build\n> echo build")

	commandCases := []struct {
		script string
		want   []string
	}{
		{"pnpm buld", []string{"pnpm build"}},
		{"pnpm buld --silent", []string{"pnpm build --silent"}},
	}
	for _, testCase := range commandCases {
		t.Run("new_command/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, pnpmMissingScriptOutput(), testCase.want)
		})
	}
}
