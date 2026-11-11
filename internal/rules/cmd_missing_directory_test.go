package rules

import "testing"

// cmdMissingDirectoryOutput is the cmd.exe error for a missing destination
// directory.
const cmdMissingDirectoryOutput = "The system cannot find the path specified.\n" +
	"        0 file(s) copied.\n"

// stubCmdDirectoryMissing replaces the directory check of the rule.
func stubCmdDirectoryMissing(t *testing.T, missing bool) {
	t.Helper()
	original := cmdDirectoryMissing
	cmdDirectoryMissing = func(string) bool { return missing }
	t.Cleanup(func() { cmdDirectoryMissing = original })
}

func TestRuleCmdMissingDirectory(t *testing.T) {
	resetTestState(t)
	rule := Get("cmd_missing_directory")
	if rule == nil {
		t.Fatal("cmd_missing_directory rule is not registered")
	}

	stubCmdDirectoryMissing(t, true)
	assertMatch(t, rule, "copy file.txt nodir\\file.txt", cmdMissingDirectoryOutput)
	assertMatch(t, rule, "move file.txt nodir\\file.txt", cmdMissingDirectoryOutput)
	assertMatch(t, rule, "copy \"my file.txt\" \"nodir\\my file.txt\"", cmdMissingDirectoryOutput)
	assertNotMatch(t, rule, "copy file.txt nodir\\file.txt", "")
	assertNotMatch(t, rule, "del file.txt", cmdMissingDirectoryOutput)
	assertNotMatch(t, rule, "copy file.txt file2.txt", cmdMissingDirectoryOutput)
	assertNotMatch(t, rule, "copy file.txt \\\\server\\share\\file.txt", cmdMissingDirectoryOutput)

	assertNewCommands(t, rule, "copy file.txt nodir\\file.txt", cmdMissingDirectoryOutput,
		[]string{"mkdir \"nodir\" && copy file.txt nodir\\file.txt"})
	assertNewCommands(t, rule, "move file.txt nodir\\file.txt", cmdMissingDirectoryOutput,
		[]string{"mkdir \"nodir\" && move file.txt nodir\\file.txt"})

	stubCmdDirectoryMissing(t, false)
	assertNotMatch(t, rule, "copy file.txt nodir\\file.txt", cmdMissingDirectoryOutput)
}
