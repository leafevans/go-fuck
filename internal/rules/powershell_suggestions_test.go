package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/types"
)

func TestPowerShellCommandIs(t *testing.T) {
	command := types.NewCommand("Get-ChildItem -Force", "")
	if !powerShellCommandIs(command, "get-childitem") {
		t.Fatal("powerShellCommandIs() is not case insensitive")
	}
	if powerShellCommandIs(command, "remove-item") {
		t.Fatal("powerShellCommandIs() matched another command")
	}
	if powerShellCommandIs(types.NewCommand("", ""), "get-childitem") {
		t.Fatal("powerShellCommandIs() matched an empty command")
	}
}

func TestPowerShellQuote(t *testing.T) {
	if got := powerShellQuote("it's"); got != "'it''s'" {
		t.Fatalf("powerShellQuote() = %q", got)
	}
}

func TestPowerShellInsertArgument(t *testing.T) {
	if got := powerShellInsertArgument("  Get-ChildItem", "-Force"); got != "Get-ChildItem -Force" {
		t.Fatalf("powerShellInsertArgument(no args) = %q", got)
	}
	if got := powerShellInsertArgument("Get-ChildItem path", "-Force"); got != "Get-ChildItem -Force path" {
		t.Fatalf("powerShellInsertArgument(args) = %q", got)
	}
}

func TestPowerShellParentDirectory(t *testing.T) {
	cases := map[string]string{
		"":              "",
		"C:\\":          "",
		"C:/":           "",
		"C:\\dir\\":     "C:\\dir",
		"C:/dir/":       "C:/dir",
		"C:\\dir\\file": "C:\\dir",
		"dir/file":      "dir",
		"file":          "",
	}
	for path, want := range cases {
		if got := powerShellParentDirectory(path); got != want {
			t.Fatalf("powerShellParentDirectory(%q) = %q, want %q", path, got, want)
		}
	}
	if got := powerShellCreatableDirectory("C:"); got != "" {
		t.Fatalf("powerShellCreatableDirectory(C:) = %q", got)
	}
}

func TestPowerShellItemHelpers(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if !powerShellItemIsDirectory(dir) || powerShellItemIsDirectory(file) {
		t.Fatal("powerShellItemIsDirectory() misclassified a path")
	}
	if !powerShellItemExists(file) || powerShellItemExists(filepath.Join(dir, "missing")) {
		t.Fatal("powerShellItemExists() misclassified a path")
	}
	names, err := powerShellDirectoryEntries(dir)
	if err != nil || len(names) != 1 || names[0] != "file.txt" {
		t.Fatalf("powerShellDirectoryEntries() = %q, %v", names, err)
	}
	if _, err := powerShellDirectoryEntries(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("powerShellDirectoryEntries() of a missing directory returned no error")
	}
}
