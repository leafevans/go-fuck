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
		{"cd zzzz", "mkdir -p zzzz && cd zzzz"},
	} {
		assertNewCommands(t, rule, testCase.script, "", []string{testCase.want})
	}
}
