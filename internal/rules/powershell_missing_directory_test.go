package rules

import "testing"

// powerShellMissingDirectoryCopyOutput is the output of Copy-Item when a part
// of the destination path does not exist.
func powerShellMissingDirectoryCopyOutput() string {
	return "cp : Could not find a part of the path 'C:\\root\\nodir\\file.txt'.\n"
}

// powerShellMissingDirectoryMoveOutput is the output of Move-Item when a part
// of the destination path does not exist. It does not name the path.
func powerShellMissingDirectoryMoveOutput() string {
	return "mv : Could not find a part of the path.\n"
}

// powerShellMissingDirectoryWriteOutput is the output of Set-Content when a part
// of the destination path does not exist.
func powerShellMissingDirectoryWriteOutput() string {
	return "Set-Content : Could not find a part of the path 'C:\\root\\nodir\\a.txt'.\n"
}

func TestRulePowerShellMissingDirectory(t *testing.T) {
	resetTestState(t)
	rule := Get("powershell_missing_directory")
	if rule == nil {
		t.Fatal("powershell_missing_directory rule is not registered")
	}

	assertMatch(t, rule, "cp file.txt nodir\\file.txt", powerShellMissingDirectoryCopyOutput())
	assertMatch(t, rule, "mv file.txt nodir\\file.txt", powerShellMissingDirectoryMoveOutput())
	assertMatch(t, rule, "Set-Content nodir\\a.txt x", powerShellMissingDirectoryWriteOutput())
	assertMatch(t, rule, "Out-File nodir\\a.txt", powerShellMissingDirectoryWriteOutput())
	assertMatch(t, rule, "New-Item nodir\\a.txt", powerShellMissingDirectoryWriteOutput())
	assertNotMatch(t, rule, "rm nodir\\file.txt", powerShellMissingDirectoryCopyOutput())
	assertNotMatch(t, rule, "cp file.txt file2.txt", "cp : Cannot find path 'C:\\file.txt' because it does not exist.\n")
	assertNotMatch(t, rule, "cp file.txt nodir\\file.txt", "")

	assertNewCommands(t, rule, "cp file.txt nodir\\file.txt", powerShellMissingDirectoryCopyOutput(),
		[]string{"mkdir -Force 'C:\\root\\nodir' && cp file.txt nodir\\file.txt"})
	assertNewCommands(t, rule, "mv file.txt nodir\\file.txt", powerShellMissingDirectoryMoveOutput(),
		[]string{"mkdir -Force 'nodir' && mv file.txt nodir\\file.txt"})
	assertNewCommands(t, rule, "Set-Content nodir\\a.txt x", powerShellMissingDirectoryWriteOutput(),
		[]string{"mkdir -Force 'C:\\root\\nodir' && Set-Content nodir\\a.txt x"})
}
