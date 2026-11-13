package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRuleCdCorrection(t *testing.T) {
	resetTestState(t)
	rule := Get("cd_correction")
	if rule == nil {
		t.Fatal("cd_correction rule is not registered")
	}

	for _, testCase := range []struct {
		script string
		output string
	}{
		{"cd foo", "cd: foo: No such file or directory"},
		{"cd foo/bar/baz", "cd: foo: No such file or directory"},
		{"cd foo/bar/baz", "cd: can't cd to foo/bar/baz"},
		{"cd /foo/bar/", "cd: The directory \"/foo/bar/\" does not exist"},
	} {
		assertMatch(t, rule, testCase.script, testCase.output)
	}
	for _, testCase := range []struct {
		script string
		output string
	}{
		{"cd foo", ""},
		{"", ""},
	} {
		assertNotMatch(t, rule, testCase.script, testCase.output)
	}
}

func TestRuleCdCorrectionNewCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("cd_correction")
	if rule == nil {
		t.Fatal("cd_correction rule is not registered")
	}

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "project", "subdir"), 0o755); err != nil {
		t.Fatalf("create test directories: %v", err)
	}

	original := cdCorrectionGetwd
	cdCorrectionGetwd = func() (string, error) { return root, nil }
	t.Cleanup(func() { cdCorrectionGetwd = original })

	for _, testCase := range []struct {
		script string
		want   string
	}{
		{"cd prject", "cd \"" + filepath.Join(root, "project") + "\""},
		// Windows also accepts forward slashes in a path.
		{"cd prject/subdr", "cd \"" + filepath.Join(root, "project", "subdir") + "\""},
		{"cd zzzz", "mkdir -p zzzz && cd zzzz"},
	} {
		assertNewCommands(t, rule, testCase.script, "", []string{testCase.want})
	}
}

// TestRuleCdCorrectionExistingFile checks that an existing file is not replaced
// by a mkdir suggestion: PowerShell reports is as "does not exist" although the
// file exists and creating a directory with the same name would fail.
func TestRuleCdCorrectionExistingFile(t *testing.T) {
	resetTestState(t)
	rule := Get("cd_correction")
	if rule == nil {
		t.Fatal("cd_correction rule is not registered")
	}
	existing := filepath.Join(t.TempDir(), "notes.txt")
	if err := os.WriteFile(existing, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := rule.GetNewCommand(testCommand("cd "+existing, "Cannot find path '"+existing+"' because it does not exist."))
	if err != nil {
		t.Fatalf("GetNewCommand() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("GetNewCommand() = %q, want no suggestion", got)
	}
}
