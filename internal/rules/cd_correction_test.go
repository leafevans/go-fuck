package rules

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestCdCorrectionHandlesDirectoryBoundaries(t *testing.T) {
	resetTestState(t)
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "project"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "file.txt"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if got := cdCorrectionSubDirs(root); !slices.Equal(got, []string{"project"}) {
		t.Fatalf("child directories = %q, want only project", got)
	}
	if got := cdCorrectionSubDirs(filepath.Join(root, "missing")); len(got) != 0 {
		t.Fatalf("missing directory children = %q", got)
	}
	original := cdCorrectionGetwd
	t.Cleanup(func() { cdCorrectionGetwd = original })
	cdCorrectionGetwd = func() (string, error) { return root, nil }
	for _, tc := range []struct{ script, want string }{
		{"cd .", "cd \"" + root + "\""},
		{"cd ..", "cd \"" + filepath.Dir(root) + "\""},
		{"cd ./prject", "cd \"" + filepath.Join(root, "project") + "\""},
	} {
		got, err := cdCorrectionNewCommand(testCommand(tc.script, ""))
		if err != nil || !slices.Equal(got, []string{tc.want}) {
			t.Errorf("correction for %q = %q, %v, want %q", tc.script, got, err, tc.want)
		}
	}
	for _, script := range []string{"cd", "cd " + string(filepath.Separator)} {
		if got, err := cdCorrectionNewCommand(testCommand(script, "")); err != nil || len(got) != 0 {
			t.Errorf("correction for %q = %q, %v, want no correction", script, got, err)
		}
	}
	wantErr := errors.New("working directory was removed")
	cdCorrectionGetwd = func() (string, error) { return "", wantErr }
	if got, err := cdCorrectionNewCommand(testCommand("cd project", "")); !errors.Is(err, wantErr) || len(got) != 0 {
		t.Fatalf("correction = %q, %v, want the original directory error", got, err)
	}
}

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
