package rules

import "testing"

// cmdMkdirExistingOutput is the cmd.exe error for an existing directory.
const cmdMkdirExistingOutput = "A subdirectory or file existing already exists.\n"

// stubCmdPathIsDirectory replaces the directory check of the rule.
func stubCmdPathIsDirectory(t *testing.T, isDirectory bool) {
	t.Helper()
	original := cmdPathIsDirectory
	cmdPathIsDirectory = func(string) bool { return isDirectory }
	t.Cleanup(func() { cmdPathIsDirectory = original })
}

func TestRuleCmdMkdirExisting(t *testing.T) {
	resetTestState(t)
	rule := Get("cmd_mkdir_existing")
	if rule == nil {
		t.Fatal("cmd_mkdir_existing rule is not registered")
	}

	stubCmdPathIsDirectory(t, true)
	assertMatch(t, rule, "mkdir existing", cmdMkdirExistingOutput)
	assertMatch(t, rule, "md existing", cmdMkdirExistingOutput)
	assertMatch(t, rule, "mkdir \"my dir\"", "A subdirectory or file my dir already exists.\n")
	assertNotMatch(t, rule, "mkdir existing", "")
	assertNotMatch(t, rule, "del existing", cmdMkdirExistingOutput)
	assertNotMatch(t, rule, "mkdir a b", "A subdirectory or file a already exists.\n")

	assertNewCommands(t, rule, "mkdir existing", cmdMkdirExistingOutput,
		[]string{"cd \"existing\""})
	assertNewCommands(t, rule, "mkdir \"my dir\"", "A subdirectory or file my dir already exists.\n",
		[]string{"cd \"my dir\""})

	stubCmdPathIsDirectory(t, false)
	assertNotMatch(t, rule, "mkdir existing", cmdMkdirExistingOutput)
}
