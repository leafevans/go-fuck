package rules

import "testing"

// stubPowerShellMissingFile replaces the file system checks of the rule.
func stubPowerShellMissingFile(t *testing.T, files []string) {
	t.Helper()
	originalExists := powerShellItemExists
	originalEntries := powerShellDirectoryEntries
	powerShellItemExists = func(path string) bool {
		for _, file := range files {
			if file == path {
				return true
			}
		}
		return false
	}
	powerShellDirectoryEntries = func(string) ([]string, error) { return files, nil }
	t.Cleanup(func() {
		powerShellItemExists = originalExists
		powerShellDirectoryEntries = originalEntries
	})
}

const powerShellMissingFileOutput = "Cannot find path '/repo/config.tx' because it does not exist.\n"
const powerShellRenameMissingOutput = "Cannot rename because item at 'config.tx' does not exist.\n"

func TestRulePowerShellMissingFile(t *testing.T) {
	resetTestState(t)
	rule := Get("powershell_missing_file")
	if rule == nil {
		t.Fatal("powershell_missing_file rule is not registered")
	}

	stubPowerShellMissingFile(t, []string{"config.txt", "my config.txt"})
	assertMatch(t, rule, "Get-Content config.tx", powerShellMissingFileOutput)
	assertMatch(t, rule, "type config.tx", powerShellMissingFileOutput)
	assertMatch(t, rule, "Copy-Item config.tx dest.txt", powerShellMissingFileOutput)
	assertMatch(t, rule, "Remove-Item config.tx", powerShellMissingFileOutput)
	assertMatch(t, rule, "Get-Item config.tx", powerShellMissingFileOutput)
	assertMatch(t, rule, "Rename-Item config.tx new.txt", powerShellRenameMissingOutput)
	assertNotMatch(t, rule, "Set-Location config.tx", powerShellMissingFileOutput)
	assertNotMatch(t, rule, "Get-Content config.txt", powerShellMissingFileOutput)
	assertNotMatch(t, rule, "Get-Content *.tx", powerShellMissingFileOutput)
	assertNotMatch(t, rule, "Get-Content config.tx", "")

	assertNewCommands(t, rule, "Get-Content config.tx", powerShellMissingFileOutput,
		[]string{"Get-Content config.txt"})
	assertNewCommands(t, rule, "type config.tx", powerShellMissingFileOutput,
		[]string{"type config.txt"})
	assertNewCommands(t, rule, "Copy-Item config.tx dest.txt", powerShellMissingFileOutput,
		[]string{"Copy-Item config.txt dest.txt"})
	assertNewCommands(t, rule, "Remove-Item config.tx", powerShellMissingFileOutput,
		[]string{"Remove-Item config.txt"})
	assertNewCommands(t, rule, "Rename-Item config.tx new.txt", powerShellRenameMissingOutput,
		[]string{"Rename-Item config.txt new.txt"})
	assertNewCommands(t, rule, "type \"my config.tx\"", powerShellMissingFileOutput,
		[]string{"type \"my config.txt\""})

	stubPowerShellMissingFile(t, []string{"config.txt"})
	assertNewCommands(t, rule, "type CONFIG.TX",
		"Cannot find path '/repo/CONFIG.TX' because it does not exist.\n",
		[]string{"type config.txt"})
}
