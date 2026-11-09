package rules

import "testing"

// powerShellMkdirExistingOutput is the output of mkdir for a directory that
// already exists.
func powerShellMkdirExistingOutput() string {
	return "mkdir : An item with the specified name C:\\root\\existing already exists.\n"
}

// stubPowerShellItemIsDirectory replaces the path check of the rule.
func stubPowerShellItemIsDirectory(t *testing.T, isDirectory bool) {
	t.Helper()
	original := powerShellItemIsDirectory
	powerShellItemIsDirectory = func(string) bool { return isDirectory }
	t.Cleanup(func() { powerShellItemIsDirectory = original })
}

func TestRulePowerShellMkdirExisting(t *testing.T) {
	resetTestState(t)
	rule := Get("powershell_mkdir_existing")
	if rule == nil {
		t.Fatal("powershell_mkdir_existing rule is not registered")
	}

	stubPowerShellItemIsDirectory(t, true)
	assertMatch(t, rule, "mkdir existing", powerShellMkdirExistingOutput())
	assertMatch(t, rule, "md existing", powerShellMkdirExistingOutput())
	assertNotMatch(t, rule, "rm existing", powerShellMkdirExistingOutput())
	assertNotMatch(t, rule, "mkdir existing", "")

	assertNewCommands(t, rule, "mkdir existing", powerShellMkdirExistingOutput(),
		[]string{"mkdir -Force existing"})

	stubPowerShellItemIsDirectory(t, false)
	assertNotMatch(t, rule, "mkdir existing", powerShellMkdirExistingOutput())
}
