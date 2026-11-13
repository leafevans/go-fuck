package rules

import "testing"

// powerShellCatDirectoryOutput is the error Windows PowerShell prints when
// Get-Content is used on a directory.
const powerShellCatDirectoryOutput = "Access to the path 'C:\\Users\\example\\project\\internal' is denied.\r\n"

func TestRulePowerShellCatDirectory(t *testing.T) {
	resetTestState(t)
	original := powerShellItemIsDirectory
	t.Cleanup(func() { powerShellItemIsDirectory = original })
	powerShellItemIsDirectory = func(path string) bool {
		return path == `C:\Users\example\project\internal`
	}

	rule := Get("powershell_cat_directory")
	if rule == nil {
		t.Fatal("powershell_cat_directory rule is not registered")
	}

	assertMatch(t, rule, "cat internal", powerShellCatDirectoryOutput)
	assertMatch(t, rule, "gc internal", powerShellCatDirectoryOutput)
	assertNotMatch(t, rule, "cat readme.txt", powerShellCatDirectoryOutput)
	assertNotMatch(t, rule, "git status", powerShellCatDirectoryOutput)
	assertNotMatch(t, rule, "cat internal", "The system cannot find the path specified.")

	assertNewCommands(t, rule, "cat internal", powerShellCatDirectoryOutput, []string{"ls internal"})
	assertNewCommands(t, rule, "Get-Content internal", powerShellCatDirectoryOutput, []string{"ls internal"})
	assertNewCommands(t, rule, "Get-Content -Path internal", powerShellCatDirectoryOutput, nil)
}
