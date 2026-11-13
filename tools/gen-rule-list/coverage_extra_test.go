package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunAlreadyUpToDateWithoutCheck(t *testing.T) {
	path := filepath.Join(t.TempDir(), "README.md")
	content := "# Doc\n\n<!-- RULES_LIST_START -->\n<!-- RULES_LIST_END -->\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run(path, false); err != nil {
		t.Fatalf("run(update) error = %v", err)
	}
	if err := run(path, false); err != nil {
		t.Fatalf("run(already up to date) error = %v", err)
	}
}

func TestRunReadError(t *testing.T) {
	dir := t.TempDir()
	if err := run(dir, false); err == nil {
		t.Fatal("run() of a directory returned no error")
	}
}

func TestRunRewriteError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "README.md")
	if err := os.WriteFile(path, []byte("no markers here\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := run(path, false)
	if err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("run() error = %v, want a missing marker error", err)
	}
}

func TestRunWriteError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "README.md")
	content := "# Doc\n\n<!-- RULES_LIST_START -->\n<!-- RULES_LIST_END -->\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o444); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o644) })

	err := run(path, false)
	if err == nil {
		if runtime.GOOS == "windows" {
			t.Fatal("run() of a read only file returned no error")
		}
		t.Skip("the read only file could be written (running as root?)")
	}
}

func TestRewriteDuplicateEndMarker(t *testing.T) {
	document := "<!-- RULES_LIST_START -->\n<!-- RULES_LIST_END -->\n<!-- RULES_LIST_END -->\n"
	_, err := rewrite(document, nil)
	if err == nil || !strings.Contains(err.Error(), "more than one") {
		t.Fatalf("rewrite() error = %v, want a duplicate end marker error", err)
	}
}
