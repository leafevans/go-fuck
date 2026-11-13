package rules

import (
	"os"
	"path/filepath"
	"testing"
)

// cmdPathNotFoundOutput is the cmd.exe error for a missing path.
const cmdPathNotFoundOutput = "The system cannot find the path specified.\n"

func TestRuleCmdCDMkdir(t *testing.T) {
	resetTestState(t)
	rule := Get("cmd_cd_mkdir")
	if rule == nil {
		t.Fatal("cmd_cd_mkdir rule is not registered")
	}

	assertMatch(t, rule, "cd C:\\missing", cmdPathNotFoundOutput)
	assertMatch(t, rule, "cd /d C:\\missing", cmdPathNotFoundOutput)
	assertMatch(t, rule, "chdir C:\\missing", cmdPathNotFoundOutput)
	assertNotMatch(t, rule, "cd C:\\missing", "")
	assertNotMatch(t, rule, "copy file.txt nodir\\file.txt", cmdPathNotFoundOutput)
	assertNotMatch(t, rule, "cd", cmdPathNotFoundOutput)
	assertNotMatch(t, rule, "cd \\\\server\\share", cmdPathNotFoundOutput)

	assertNewCommands(t, rule, "cd C:\\missing", cmdPathNotFoundOutput,
		[]string{"mkdir \"C:\\missing\" && cd /d \"C:\\missing\""})
	assertNewCommands(t, rule, "cd /d C:\\missing", cmdPathNotFoundOutput,
		[]string{"mkdir \"C:\\missing\" && cd /d \"C:\\missing\""})
}

// TestRuleCmdCDMkdirExistingPath checks that an existing file is not treated as
// a missing directory.
func TestRuleCmdCDMkdirExistingPath(t *testing.T) {
	resetTestState(t)
	rule := Get("cmd_cd_mkdir")
	if rule == nil {
		t.Fatal("cmd_cd_mkdir rule is not registered")
	}
	existing := filepath.Join(t.TempDir(), "notes.txt")
	if err := os.WriteFile(existing, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	if rule.Match(testCommand("cd "+existing, cmdPathNotFoundOutput)) {
		t.Fatalf("cmd_cd_mkdir matched an existing file %q", existing)
	}
	missing := filepath.Join(t.TempDir(), "newdir")
	assertMatch(t, rule, "cd "+missing, cmdPathNotFoundOutput)
}
