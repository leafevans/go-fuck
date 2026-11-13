package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCmdInsertArgument(t *testing.T) {
	if got := cmdInsertArgument("  dir", "/b"); got != "dir /b" {
		t.Fatalf("cmdInsertArgument(no args) = %q", got)
	}
	if got := cmdInsertArgument("dir file.txt", "/b"); got != "dir /b file.txt" {
		t.Fatalf("cmdInsertArgument(with args) = %q", got)
	}
}

func TestCmdQuote(t *testing.T) {
	if got := cmdQuote("my file"); got != "\"my file\"" {
		t.Fatalf("cmdQuote() = %q", got)
	}
}

func TestCmdSplit(t *testing.T) {
	got := cmdSplit("  copy \"my file.txt\"  C:\\dest\\  ")
	want := []string{"copy", "my file.txt", "C:\\dest\\"}
	if len(got) != len(want) {
		t.Fatalf("cmdSplit() = %q, want %q", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("cmdSplit() = %q, want %q", got, want)
		}
	}
	if got := cmdSplit(""); len(got) != 0 {
		t.Fatalf("cmdSplit(empty) = %q", got)
	}
}

func TestCmdParentDirectory(t *testing.T) {
	cases := map[string]string{
		"":                    "",
		"C:\\":                "",
		"C:/":                 "",
		"C:\\dir\\":           "C:\\dir",
		"C:/dir/":             "C:/dir",
		"C:\\dir\\file":       "C:\\dir",
		"dir/file":            "dir",
		"file":                "",
		"\\\\server\\share\\": "\\\\server\\share",
	}
	for path, want := range cases {
		if got := cmdParentDirectory(path); got != want {
			t.Fatalf("cmdParentDirectory(%q) = %q, want %q", path, got, want)
		}
	}
	if got := cmdCreatableDirectory("C:"); got != "" {
		t.Fatalf("cmdCreatableDirectory(C:) = %q", got)
	}
	if got := cmdCreatableDirectory("C:\\dir"); got != "C:\\dir" {
		t.Fatalf("cmdCreatableDirectory(dir) = %q", got)
	}
}

func TestCmdFirstRawArgument(t *testing.T) {
	raw, value := cmdFirstRawArgument("copy /b \"my file.txt\" dest")
	if raw != "\"my file.txt\"" || value != "my file.txt" {
		t.Fatalf("cmdFirstRawArgument(quoted) = %q, %q", raw, value)
	}
	raw, value = cmdFirstRawArgument("copy -y file.txt")
	if raw != "file.txt" || value != "file.txt" {
		t.Fatalf("cmdFirstRawArgument(flag) = %q, %q", raw, value)
	}
	if raw, value := cmdFirstRawArgument("copy"); raw != "" || value != "" {
		t.Fatalf("cmdFirstRawArgument(no args) = %q, %q", raw, value)
	}
	if raw, value := cmdFirstRawArgument("copy \"unterminated"); raw != "\"unterminated" || value != "unterminated" {
		t.Fatalf("cmdFirstRawArgument(unterminated) = %q, %q", raw, value)
	}
}

func TestCmdClosestFile(t *testing.T) {
	original := cmdDirectoryEntries
	t.Cleanup(func() { cmdDirectoryEntries = original })
	cmdDirectoryEntries = func(dir string) ([]string, error) {
		normalized := filepath.ToSlash(dir)
		switch {
		case normalized == "C:/proj" || strings.HasSuffix(normalized, "/proj"):
			return []string{"App.js", "app.js", "other.js"}, nil
		case dir == ".":
			return []string{"main.go"}, nil
		}
		return nil, os.ErrNotExist
	}

	if got := cmdClosestFile("C:/proj/APP.JS"); got != filepath.Join("C:/proj", "App.js") {
		t.Fatalf("cmdClosestFile(casefold) = %q", got)
	}
	if got := cmdClosestFile("C:/proj/appp.js"); got != filepath.Join("C:/proj", "app.js") {
		t.Fatalf("cmdClosestFile(closest) = %q", got)
	}
	if got := cmdClosestFile("C:/proj/nothing.txt"); got != "" {
		t.Fatalf("cmdClosestFile(none) = %q", got)
	}
	if got := cmdClosestFile("missing/nothing"); got != "" {
		t.Fatalf("cmdClosestFile(missing dir) = %q", got)
	}
}

func TestCmdReplaceRawArgument(t *testing.T) {
	if got := cmdReplaceRawArgument("copy file.txt dest", "file.txt", "new.txt"); got != "copy new.txt dest" {
		t.Fatalf("cmdReplaceRawArgument() = %q", got)
	}
	if got := cmdReplaceRawArgument("copy file.txt dest", "file.txt", "new file.txt"); got != "copy \"new file.txt\" dest" {
		t.Fatalf("cmdReplaceRawArgument(quote) = %q", got)
	}
}

func TestCmdFirstWord(t *testing.T) {
	if got := cmdFirstWord("  copy file.txt"); got != "copy" {
		t.Fatalf("cmdFirstWord() = %q", got)
	}
	if got := cmdFirstWord("copy"); got != "copy" {
		t.Fatalf("cmdFirstWord(single) = %q", got)
	}
}

func TestCmdPathHelpers(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if !cmdPathIsDirectory(dir) || cmdPathIsDirectory(file) {
		t.Fatal("cmdPathIsDirectory() misclassified a path")
	}
	if !cmdFileExists(file) || cmdFileExists(filepath.Join(dir, "missing")) {
		t.Fatal("cmdFileExists() misclassified a path")
	}
	if !cmdDirectoryMissing(filepath.Join(dir, "missing")) || cmdDirectoryMissing(dir) {
		t.Fatal("cmdDirectoryMissing() misclassified a path")
	}
	names, err := cmdDirectoryEntries(dir)
	if err != nil || len(names) != 1 || names[0] != "file.txt" {
		t.Fatalf("cmdDirectoryEntries() = %q, %v", names, err)
	}
	if _, err := cmdDirectoryEntries(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("cmdDirectoryEntries() of a missing directory returned no error")
	}
}
