package rules

import (
	"slices"
	"testing"
)

// stubPowerShellMissingFile replaces the file system checks of the rule.
func stubPowerShellMissingFile(t *testing.T, files []string) {
	t.Helper()
	originalExists := powerShellItemExists
	originalEntries := powerShellDirectoryEntries
	powerShellItemExists = func(path string) bool { return slices.Contains(files, path) }
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

func TestRulePowerShellMissingFileDeclinesUnsafeSubstitutions(t *testing.T) {
	resetTestState(t)
	rule := Get("powershell_missing_file")
	for _, candidates := range [][]string{
		{"config.ta", "config.tb"},
		{"Config.txt", "config.txt"},
		{"config tx"},
		{"config'tx"},
		{"config[tx]"},
	} {
		stubPowerShellMissingFile(t, candidates)
		assertNewCommands(t, rule, "Rename-Item config.tx new.txt", powerShellRenameMissingOutput, nil)
	}
	stubPowerShellMissingFile(t, []string{"config.txt"})
	assertNotMatch(t, rule, "Rename-Item config.tx config.tx", powerShellRenameMissingOutput)
	assertNewCommands(t, rule, "Rename-Item config.tx config.tx", powerShellRenameMissingOutput, nil)
	assertNotMatch(t, rule, "Get-Content config[1].tx", "Cannot find path 'config[1].tx' because it does not exist.")
}

func TestRulePowerShellMissingFileChinese(t *testing.T) {
	resetTestState(t)
	stubPowerShellMissingFile(t, []string{"config.txt"})
	rule := Get("powershell_missing_file")
	readOutput := "Get-Content: \u627e\u4e0d\u5230\u8def\u5f84\u201cconfig.tx\u201d\uff0c\u56e0\u4e3a\u8be5\u8def\u5f84\u4e0d\u5b58\u5728\u3002\r\n"
	renameOutput := "Rename-Item: \u65e0\u6cd5\u91cd\u547d\u540d\uff0c\u56e0\u4e3a\u201cconfig.tx\u201d\u5904\u4e0d\u5b58\u5728\u9879\u3002\r\n"
	read := "Get-Content -LiteralPath 'config.tx'"
	rename := "Rename-Item -LiteralPath 'config.tx' -NewName 'new.txt'"
	assertMatch(t, rule, read, readOutput)
	assertNewCommands(t, rule, read, readOutput, []string{"Get-Content -LiteralPath 'config.txt'"})
	assertMatch(t, rule, rename, renameOutput)
	assertNewCommands(t, rule, rename, renameOutput, []string{"Rename-Item -LiteralPath 'config.txt' -NewName 'new.txt'"})
	altRenameOutput := "Rename-Item: \u65e0\u6cd5\u91cd\u547d\u540d\uff0c\u56e0\u4e3a\u201cconfig.tx\u201d\u5904\u7684\u9879\u4e0d\u5b58\u5728\u3002\r\n"
	assertMatch(t, rule, rename, altRenameOutput)
	assertNewCommands(t, rule, rename, altRenameOutput, []string{"Rename-Item -LiteralPath 'config.txt' -NewName 'new.txt'"})
	assertNotMatch(t, rule, read, "Get-Content: \u62d2\u7edd\u8bbf\u95ee\r\n")
	assertNotMatch(t, rule, rename, "Rename-Item: \u76ee\u6807\u5df2\u5b58\u5728\r\n")
}

// TestRulePowerShellMissingFileListing checks that listing commands (ls, dir,
// gci) also get the closest existing name suggested.
func TestRulePowerShellMissingFileListing(t *testing.T) {
	resetTestState(t)
	rule := Get("powershell_missing_file")
	if rule == nil {
		t.Fatal("powershell_missing_file rule is not registered")
	}
	if rule.Priority != 900 {
		t.Errorf("powershell_missing_file priority = %d, want 900", rule.Priority)
	}
	stubPowerShellMissingFile(t, []string{"config.txt"})
	output := "Cannot find path '/repo/config.tx' because it does not exist.\n"
	assertMatch(t, rule, "ls config.tx", output)
	assertMatch(t, rule, "dir config.tx", output)
	assertMatch(t, rule, "gci config.tx", output)
	assertMatch(t, rule, "Get-ChildItem config.tx", output)
	assertNewCommands(t, rule, "ls config.tx", output, []string{"ls config.txt"})
	assertNewCommands(t, rule, "Get-ChildItem config.tx", output, []string{"Get-ChildItem config.txt"})
}
