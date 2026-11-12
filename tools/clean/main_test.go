package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRemovesPaths(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "artifact.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	subdir := filepath.Join(dir, "sub")
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatal(err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := run([]string{file, subdir}, stdout, stderr)
	if code != 0 {
		t.Fatalf("run() = %d, stderr = %q", code, stderr.String())
	}
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Fatalf("file still exists: %v", err)
	}
	if _, err := os.Stat(subdir); !os.IsNotExist(err) {
		t.Fatalf("directory still exists: %v", err)
	}
	if !strings.Contains(stdout.String(), "removed") {
		t.Fatalf("stdout = %q, want removed lines", stdout.String())
	}
}

func TestRunMissingPath(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := run([]string{filepath.Join(t.TempDir(), "missing")}, stdout, stderr)
	if code != 0 {
		t.Fatalf("run() = %d, want 0 for a missing path", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want no output", stdout.String())
	}
}

func TestRunWithoutArguments(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := run(nil, stdout, stderr)
	if code != 2 {
		t.Fatalf("run() = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "usage") {
		t.Fatalf("stderr = %q, want usage", stderr.String())
	}
}
