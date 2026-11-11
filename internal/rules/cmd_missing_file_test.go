package rules

import "testing"

// stubCmdMissingFile replaces the file system checks of the rule.
func stubCmdMissingFile(t *testing.T, files []string) {
	t.Helper()
	originalExists := cmdFileExists
	originalEntries := cmdDirectoryEntries
	cmdFileExists = func(path string) bool {
		for _, file := range files {
			if file == path {
				return true
			}
		}
		return false
	}
	cmdDirectoryEntries = func(string) ([]string, error) { return files, nil }
	t.Cleanup(func() {
		cmdFileExists = originalExists
		cmdDirectoryEntries = originalEntries
	})
}

const cmdFileNotFoundOutput = "The system cannot find the file specified.\n"

func TestRuleCmdMissingFile(t *testing.T) {
	resetTestState(t)
	rule := Get("cmd_missing_file")
	if rule == nil {
		t.Fatal("cmd_missing_file rule is not registered")
	}

	stubCmdMissingFile(t, []string{"config.txt", "my config.txt"})
	assertMatch(t, rule, "type config.tx", cmdFileNotFoundOutput)
	assertMatch(t, rule, "copy config.tx dest.txt", cmdFileNotFoundOutput)
	assertMatch(t, rule, "del config.tx", "Could Not Find C:\\root\\config.tx\n")
	assertMatch(t, rule, "ren config.tx new.txt", cmdFileNotFoundOutput)
	assertMatch(t, rule, "type \"my config.tx\"", cmdFileNotFoundOutput)
	assertNotMatch(t, rule, "type config.txt", cmdFileNotFoundOutput)
	assertNotMatch(t, rule, "where config.tx", cmdFileNotFoundOutput)
	assertNotMatch(t, rule, "type config.tx", "The system cannot find the path specified.\n")

	assertNewCommands(t, rule, "type config.tx", cmdFileNotFoundOutput,
		[]string{"type config.txt"})
	assertNewCommands(t, rule, "copy config.tx dest.txt", cmdFileNotFoundOutput,
		[]string{"copy config.txt dest.txt"})
	assertNewCommands(t, rule, "ren config.tx new.txt", cmdFileNotFoundOutput,
		[]string{"ren config.txt new.txt"})
	assertNewCommands(t, rule, "type \"my config.tx\"", cmdFileNotFoundOutput,
		[]string{"type \"my config.txt\""})
}
