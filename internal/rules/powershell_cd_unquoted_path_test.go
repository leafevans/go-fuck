package rules

import "testing"

// powerShellCDUnquotedPathOutput is the output of a directory change with an
// unquoted path that contains a space.
func powerShellCDUnquotedPathOutput() string {
	return "Set-Location : A positional parameter cannot be found that accepts argument 'Files'.\n"
}

func TestRulePowerShellCDUnquotedPath(t *testing.T) {
	resetTestState(t)
	rule := Get("powershell_cd_unquoted_path")
	if rule == nil {
		t.Fatal("powershell_cd_unquoted_path rule is not registered")
	}

	assertMatch(t, rule, "cd C:\\Program Files", powerShellCDUnquotedPathOutput())
	assertMatch(t, rule, "Set-Location C:\\Program Files", powerShellCDUnquotedPathOutput())
	assertNotMatch(t, rule, "rm C:\\Program Files", powerShellCDUnquotedPathOutput())
	assertNotMatch(t, rule, "cd C:\\Program", "")

	assertNewCommands(t, rule, "cd C:\\Program Files", powerShellCDUnquotedPathOutput(),
		[]string{"cd 'C:\\Program Files'"})
	assertNewCommands(t, rule, "Set-Location C:\\Program Files", powerShellCDUnquotedPathOutput(),
		[]string{"Set-Location 'C:\\Program Files'"})
}
