package rules

import (
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v3/internal/shells"
)

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

// TestRulePowerShellCatDirectoryAbsolutePath covers absolute Windows paths.
//
// The PowerShell shell keeps the backslashes when it splits the command line,
// so unlike the POSIX splitting the path of the command matches the path of the
// error message.
func TestRulePowerShellCatDirectoryAbsolutePath(t *testing.T) {
	resetTestState(t)
	shells.Set(shells.NewPowershell())
	t.Cleanup(func() { shells.Set(shells.NewGeneric()) })
	original := powerShellItemIsDirectory
	t.Cleanup(func() { powerShellItemIsDirectory = original })
	powerShellItemIsDirectory = func(path string) bool {
		return strings.EqualFold(path, `C:\Windows`)
	}

	rule := Get("powershell_cat_directory")
	if rule == nil {
		t.Fatal("powershell_cat_directory rule is not registered")
	}
	output := "Access to the path 'C:\\Windows' is denied.\r\n"
	assertMatch(t, rule, `cat C:\Windows`, output)
	assertMatch(t, rule, `gc "C:\Windows"`, output)
	assertNotMatch(t, rule, `cat C:\Windows2`, output)
	assertNewCommands(t, rule, `cat C:\Windows`, output,
		[]string{`ls C:\Windows`})
}
