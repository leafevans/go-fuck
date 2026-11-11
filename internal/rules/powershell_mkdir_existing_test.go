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
	assertMatch(t, rule, "New-Item -ItemType Directory existing", powerShellMkdirExistingOutput())
	assertMatch(t, rule, "New-Item -Type Directory existing", powerShellMkdirExistingOutput())
	assertNotMatch(t, rule, "rm existing", powerShellMkdirExistingOutput())
	assertNotMatch(t, rule, "mkdir existing", "")
	assertNotMatch(t, rule, "New-Item existing", powerShellMkdirExistingOutput())
	assertNotMatch(t, rule, "New-Item -ItemType File existing", powerShellMkdirExistingOutput())

	assertNewCommands(t, rule, "mkdir existing", powerShellMkdirExistingOutput(),
		[]string{"mkdir -Force existing"})
	assertNewCommands(t, rule, "New-Item -ItemType Directory existing", powerShellMkdirExistingOutput(),
		[]string{"New-Item -Force -ItemType Directory existing"})

	stubPowerShellItemIsDirectory(t, false)
	assertNotMatch(t, rule, "mkdir existing", powerShellMkdirExistingOutput())
}
