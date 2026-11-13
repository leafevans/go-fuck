package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRuleCdMkdir(t *testing.T) {
	rule := Get("cd_mkdir")
	if rule == nil {
		t.Fatal("cd_mkdir rule is not registered")
	}

	for _, script := range []string{
		"cd foo",
		"cd foo/bar",
	} {
		assertMatch(t, rule, script, "cd: foo: No such file or directory")
	}
	assertMatch(t, rule, "cd foo", "cd: can't cd to foo")
	assertMatch(t, rule, "cd /foo/bar", "cd: The directory \"/foo/bar\" does not exist")

	assertNotMatch(t, rule, "cd foo", "")
	assertNotMatch(t, rule, "ls foo", "cd: foo: No such file or directory")

	assertNewCommands(t, rule, "cd foo", "cd: foo: No such file or directory",
		[]string{"mkdir -p foo && cd foo"})
	assertNewCommands(t, rule, "sudo cd foo", "cd: foo: No such file or directory",
		[]string{"sudo mkdir -p foo && cd foo"})
}

// TestRuleCdMkdirExistingPath checks that an existing file is not treated as a
// missing directory: PowerShell reports a file as "does not exist" and creating
// a directory with the same name would fail.
func TestRuleCdMkdirExistingPath(t *testing.T) {
	resetTestState(t)
	rule := Get("cd_mkdir")
	if rule == nil {
		t.Fatal("cd_mkdir rule is not registered")
	}
	existing := filepath.Join(t.TempDir(), "notes.txt")
	if err := os.WriteFile(existing, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	if rule.Match(testCommand("cd "+existing, "Cannot find path '"+existing+"' because it does not exist.")) {
		t.Fatalf("cd_mkdir matched an existing file %q", existing)
	}
	missing := filepath.Join(t.TempDir(), "newdir")
	assertMatch(t, rule, "cd "+missing, "Cannot find path '"+missing+"' because it does not exist.")
}
