package rules

import "testing"

// powerShellRemoveRecursePromptOutput is the interactive prompt that asks for
// the Recurse parameter when a directory with children is removed.
func powerShellRemoveRecursePromptOutput() string {
	return "Confirm\n" +
		"The item at C:\\root\\dir has children and the Recurse parameter was not specified. " +
		"If you continue, all children will be removed with the item. Are you sure you want to continue?\n"
}

// powerShellRemoveRecurseErrorOutput is the non-interactive error that
// PowerShell prints when the confirmation prompt cannot be shown.
func powerShellRemoveRecurseErrorOutput() string {
	return "rm : Windows PowerShell is in NonInteractive mode. Read and Prompt functionality is not available.\n"
}

func TestRulePowerShellRemoveRecurse(t *testing.T) {
	resetTestState(t)
	rule := Get("powershell_remove_recurse")
	if rule == nil {
		t.Fatal("powershell_remove_recurse rule is not registered")
	}

	assertMatch(t, rule, "rm dir", powerShellRemoveRecursePromptOutput())
	assertMatch(t, rule, "rm dir", powerShellRemoveRecurseErrorOutput())
	assertMatch(t, rule, "Remove-Item dir", powerShellRemoveRecurseErrorOutput())
	assertMatch(t, rule, "del dir", powerShellRemoveRecurseErrorOutput())
	assertNotMatch(t, rule, "cp dir backup", powerShellRemoveRecurseErrorOutput())
	assertNotMatch(t, rule, "rm file.txt", "rm : Cannot find path 'C:\\file.txt' because it does not exist.\n")
	assertNotMatch(t, rule, "rm dir", "")

	assertNewCommands(t, rule, "rm dir", powerShellRemoveRecurseErrorOutput(),
		[]string{"rm -Recurse dir"})
	assertNewCommands(t, rule, "Remove-Item dir", powerShellRemoveRecurseErrorOutput(),
		[]string{"Remove-Item -Recurse dir"})
}
